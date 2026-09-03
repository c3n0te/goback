package main

import (
	"log/slog"
	"net/http"
)

func newRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items", getItems)
	mux.HandleFunc("GET /items/{id}", getItemById)
	mux.HandleFunc("POST /items", createItem)
	return mux
}

func main() {
	mux := newRouter()

	slog.Info("Server running on port :8000")
	if err := http.ListenAndServe(":8000", mux); err != nil {
		slog.Error("Failed to listen on port 8000")
		panic(err)
	}

}
