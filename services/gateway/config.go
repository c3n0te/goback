package main

import (
	"os"
)

type Config struct {
	GrpcIp   string
	GrpcPort string
	HttpIp   string
	HttpPort string
}

func NewConfig() *Config {
	grpcAddr := os.Getenv("GRPC_IP")
	if grpcAddr == "" {
		grpcAddr = "txserver"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	addr := os.Getenv("HTTP_IP")
	if addr == "" {
		addr = "gateway"
	}

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8000"
	}

	cfg := Config{
		GrpcIp:   grpcAddr,
		GrpcPort: grpcPort,
		HttpIp:   addr,
		HttpPort: port,
	}

	return &cfg
}
