package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Abhishek48Shah/bot/internal/api"
	"github.com/Abhishek48Shah/bot/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func loadDatabase(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return pool, nil
}
func main() {

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	pool, err := loadDatabase(ctx, "postgres://chatbot:chatbot123@localhost:5432/chat_bot?sslmode=disable")
	fmt.Printf("database connection established: port: %d\n", 5432)
	if err != nil {
		logger.Error("failed to connect postgres", "error", err)
		os.Exit(1)
	}
	middleware := middleware.NewMiddleware(logger)
	server := &http.Server{Addr: ":4000", Handler: api.Api(pool, logger, middleware)}
	logger.Info("server is running", "port", server.Addr)
	err = server.ListenAndServe()
	fmt.Errorf("failed to run server: %w", err)
}
