package main

import "os"

type Config struct {
	TcpIp   string
	TcpPort string
}

func NewConfig() *Config {
	tcpIp := os.Getenv("TCP_IP")
	if tcpIp == "" {
		tcpIp = "quote"
	}

	tcpPort := os.Getenv("TCP_PORT")
	if tcpPort == "" {
		tcpPort = "4444"
	}

	cfg := &Config{
		TcpIp:   tcpIp,
		TcpPort: tcpPort,
	}

	return cfg
}
