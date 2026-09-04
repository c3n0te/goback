package main

import (
	"api"
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc/metadata"
)

func CallRegister(gbc api.GatewayClient, cfg *Config) (*api.RegResponse, error) {
	md := metadata.Pairs("timestamp", time.Now().UTC().Format(time.StampNano))
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	regreq := &api.RegRequest{
		DeviceId: "1",
	}

	resp, err := gbc.Register(ctx, regreq)
	if err != nil {
		slog.Error("Failed to register gateway gRPC client", "error", err)
		return nil, err
	}

	slog.Info(fmt.Sprintf("Connected %v", resp.Registered))
	return resp, nil
}
