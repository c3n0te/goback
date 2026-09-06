package main

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

var src = rand.NewSource(time.Now().UnixNano())

const (
	characters    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func GenerateQuote() (float64, int64, string) {
	return rand.Float64() * 100, time.Now().Unix(), GenerateCryptoKey(44)
}

func GenerateCryptoKey(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(characters) {
			b[i] = characters[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func HandleClient(conn net.Conn) {
	defer conn.Close()
	slog.Info(fmt.Sprintf("New client connected: %s", conn.RemoteAddr().String()))
	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				slog.Error(fmt.Sprintf("Client disconnected: %s", conn.RemoteAddr().String()))
			} else {
				slog.Error("TCP Read error: ", "error", err)
			}
			return
		}

		ticker := strings.TrimSpace(msg)
		price, ts, key := GenerateQuote()
		_, err = conn.Write([]byte(fmt.Sprintf("%v,%v,%v,%v\n", ticker, price, ts, key)))
		if err != nil {
			slog.Error("TCP Write error: ", "error", err)
			return
		}
	}
}

func main() {
	cfg := NewConfig()

	tcpAddr := fmt.Sprintf("%v:%v", cfg.TcpIp, cfg.TcpPort)
	listener, err := net.Listen("tcp", tcpAddr)
	if err != nil {
		slog.Error("Error starting TCP server: ", "error", err)
		os.Exit(1)
	}
	defer listener.Close()

	slog.Info(fmt.Sprintf("TCP Server listening on: %v", tcpAddr))
	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("Error accepting connection: ", "error", err)
			continue
		}

		go HandleClient(conn)
	}
}
