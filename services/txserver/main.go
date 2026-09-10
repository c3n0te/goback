package main

import (
	"api"
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func NewDbWithRetry(cfg *Config) *sqlx.DB {
	var db *sqlx.DB
	var err error

	slog.Info(fmt.Sprintf("DB URL: %v", cfg.DbUrl))
	for range 5 {
		db, err = sqlx.Connect(cfg.DbType, cfg.DbUrl)
		if err != nil {
			slog.Error("Failed to connect to DB: ", "error", err)
			slog.Info("DB Not ready, sleeping for 3 seconds")
			time.Sleep(3 * time.Second)
			continue
		}

		break
	}

	if db == nil {
		slog.Error("Failed to connect to DB")
		os.Exit(1)
	}

	slog.Info("DB ready")
	return db
}

func NewCacheWithRetry(cfg *Config) *redis.Client {
	addr := fmt.Sprintf("%s:%s", cfg.CacheIp, cfg.CachePort)
	slog.Info(fmt.Sprintf("Attempting to connect to cache at address: %v", addr))
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
		ClientSideCacheConfig: &redis.ClientSideCacheConfig{
			MaxEntries: 10_000,
		},
	})

	ctx := context.Background()
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		slog.Error("Could not connect to Redis: ", "error", err)
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("Cache ready: %v", pong))
	return rdb
}

func NewQueueWithRetry(cfg *Config) (*rmq.AmqpConnection, *rmq.Environment, *rmq.Consumer) {
	ctx := context.Background()
	env := rmq.NewEnvironment(cfg.QueueUrl, nil)
	var conn *rmq.AmqpConnection
	var consumer *rmq.Consumer
	var err error

	for range 5 {
		conn, err = env.NewConnection(ctx)
		if err != nil {
			slog.Error("Failed to connect to RabbitMQ: ", "error", err)
			slog.Info("Queue Not ready, sleeping for 3 seconds")
			time.Sleep(3 * time.Second)
			continue
		}

		break
	}

	_, err = conn.Management().DeclareQueue(ctx, &rmq.QuorumQueueSpecification{
		Name: cfg.QueueName,
	})

	if err != nil {
		slog.Error("Failed to declare a queue: ", "error", err)

	}

	consumer, err = conn.NewConsumer(ctx, cfg.QueueName, nil)
	if err != nil {
		slog.Error("Failed to create queue consumer: ", "error", err)
	}

	if conn == nil || env == nil || consumer == nil {
		slog.Error("Failed to connect to Queue")
		os.Exit(1)
	}

	slog.Info("Queue ready")
	return conn, env, consumer
}

func NewQuoteConnWithRetry(cfg *Config) net.Conn {
	var quoteConn net.Conn
	var err error

	tcpAddr := fmt.Sprintf("%v:%v", cfg.QuoteIp, cfg.QuotePort)
	slog.Info(fmt.Sprintf("Quote Server Address: %v", tcpAddr))
	for range 5 {
		quoteConn, err = net.Dial("tcp", tcpAddr)
		if err != nil {
			slog.Error("Failed to connect to Quote Server: ", "error", err)
			slog.Info("Quote Server Not ready, sleeping for 3 seconds")
			time.Sleep(3 * time.Second)
			continue
		}

		break
	}

	if quoteConn == nil {
		slog.Error("Failed to connect to Quote Server")
		os.Exit(1)
	}

	slog.Info("Quote Server ready")
	return quoteConn
}

func main() {
	cfg := NewConfig()

	db := NewDbWithRetry(cfg)
	defer db.Close()
	Migrate(db)

	cache := NewCacheWithRetry(cfg)
	defer cache.Close()

	_, env, qconsumer := NewQueueWithRetry(cfg)
	defer func() {
		_ = env.CloseConnections(context.Background())
		_ = qconsumer.Close(context.Background())
	}()

	quoteConn := NewQuoteConnWithRetry(cfg)
	defer quoteConn.Close()

	srv := grpc.NewServer()
	goback := NewGoBackServer(db, cache, qconsumer, quoteConn)
	go goback.ConsumeQueue()
	api.RegisterGoBackServer(srv, &goback)
	grpcAddr := fmt.Sprintf("%v:%v", cfg.GrpcIp, cfg.GrpcPort)
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		slog.Error("Failed to listen to socket: ", "error", err)
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("gRPC Server running on %v", grpcAddr))
	err = srv.Serve(listener)
	if err != nil {
		slog.Error("Failed to bind server to listener: ", "error", err)
		os.Exit(1)
	}
}
