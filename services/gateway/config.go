package main

import (
	"os"
)

type Config struct {
	GrpcIp   string
	GrpcPort string
	HttpIp   string
	HttpPort string
	DbType   string
	DbHost   string
	DbPort   string
	DbUser   string
	DbPass   string
	DbName   string
	DbSsl    string
}

func NewConfig() *Config {
	grpcAddr := os.Getenv("GRPC_IP")
	if grpcAddr == "" {
		grpcAddr = "127.0.0.1"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	addr := os.Getenv("HTTP_IP")
	if addr == "" {
		addr = "127.0.0.1"
	}

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8000"
	}

	db_type := os.Getenv("DB_TYPE")
	if db_type == "" {
		db_type = "postgres"
	}

	db_host := os.Getenv("DB_HOST")
	if db_host == "" {
		db_host = "127.0.0.1"
	}

	db_port := os.Getenv("DB_PORT")
	if db_port == "" {
		db_port = "5432"
	}

	db_user := os.Getenv("DB_USER")
	if db_user == "" {
		db_user = "postgres"
	}

	db_pass := os.Getenv("DB_PASS")
	if db_pass == "" {
		db_pass = "postgres"
	}

	db_name := os.Getenv("DB_NAME")
	if db_name == "" {
		db_name = "postgres"
	}

	db_ssl := os.Getenv("DB_SSL")
	if db_ssl == "" {
		db_ssl = "disable"
	}

	cfg := Config{
		GrpcIp:   grpcAddr,
		GrpcPort: grpcPort,
		HttpIp:   addr,
		HttpPort: port,
		DbType:   db_type,
		DbHost:   db_host,
		DbPort:   db_port,
		DbUser:   db_user,
		DbPass:   db_pass,
		DbName:   db_name,
		DbSsl:    db_ssl,
	}

	return &cfg
}
