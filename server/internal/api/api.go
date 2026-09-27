package api

import (
	"log/slog"
	"net/http"

	"github.com/Abhishek48Shah/bot/internal/bot"
	"github.com/Abhishek48Shah/bot/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Api(pool *pgxpool.Pool, logger *slog.Logger, middleware *middleware.Middleware) http.Handler {
	mux := http.NewServeMux()
	botStorage := bot.NewStorage(pool)
	ollamaClient := bot.NewOllamaClient(11434, "localhost", "gemma3:4b")
	embded := bot.NewEmbedding("localhost", 11434, "qwen3-embedding:0.6b")
	botService := bot.NewService(botStorage, embded, ollamaClient)
	botHandler := bot.NewHandler(botService, logger)
	mux.HandleFunc("POST /api/chat/stream", botHandler.HandleChatStream)
	return middleware.PanicRecovery(middleware.CorsMiddleware(mux))
}
