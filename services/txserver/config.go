package main

import (
	"os"
)

type Config struct {
	DbUrl     string
	DbType    string
	GrpcIp    string
	GrpcPort  string
	CacheIp   string
	CachePort string
	QueueUrl  string
}

func NewConfig() *Config {
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		dbUrl = "postgresql://root@cockroach:26257/defaultdb?sslmode=disable"
	}

	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "postgres"
	}

	grpcAddr := os.Getenv("GRPC_IP")
	if grpcAddr == "" {
		grpcAddr = "txserver"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	cacheIp := os.Getenv("CACHE_IP")
	if cacheIp == "" {
		cacheIp = "redis"
	}

	cachePort := os.Getenv("CACHE_PORT")
	if cachePort == "" {
		cachePort = "6379"
	}

	queueUrl := os.Getenv("QUEUE_URL")
	if queueUrl == "" {
		queueUrl = "amqp://guest:guest@rmq:5672/"
	}

	cfg := Config{
		DbUrl:     dbUrl,
		DbType:    dbType,
		GrpcIp:    grpcAddr,
		GrpcPort:  grpcPort,
		CacheIp:   cacheIp,
		CachePort: cachePort,
		QueueUrl:  queueUrl,
	}

	return &cfg
}
