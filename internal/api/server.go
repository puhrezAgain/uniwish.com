package api

import (
	"log/slog"
	"net/http"

	"uniwish.com/internal/api/middleware"
)

func NewServer(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	RegisterRoutes(mux)

	handler := middleware.Logging(logger)(mux)
	return handler
}
