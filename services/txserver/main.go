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
			slog.Error("Failed to connect to DB", "error", err)
			slog.Info("DB Not ready, sleeping for 3 seconds")
			time.Sleep(3 * time.Second)
			continue
		}

		slog.Info("DB ready")
		break
	}
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
	}

	slog.Info(fmt.Sprintf("Cache ready: %v", pong))
	return rdb
}

func NewQueueWithRetry(cfg *Config) (*rmq.AmqpConnection, *rmq.Environment) {
	ctx := context.Background()
	env := rmq.NewEnvironment(cfg.QueueUrl, nil)
	var conn *rmq.AmqpConnection
	var err error

	for range 5 {
		conn, err = env.NewConnection(ctx)
		if err != nil {
			slog.Error("Failed to connect to RabbitMQ: %v", "error", err)
			slog.Info("Queue Not ready, sleeping for 3 seconds")
			time.Sleep(3 * time.Second)
			continue
		}

		slog.Info("Queue ready")
		break
	}

	return conn, env
}

func main() {
	cfg := NewConfig()

	db := NewDbWithRetry(cfg)
	defer db.Close()
	Migrate(db)

	cache := NewCacheWithRetry(cfg)
	defer cache.Close()

	queue, env := NewQueueWithRetry(cfg)
	defer func() {
		_ = env.CloseConnections(context.Background())
	}()

	srv := grpc.NewServer()
	goback := NewGoBackServer(db, cache, queue)
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
