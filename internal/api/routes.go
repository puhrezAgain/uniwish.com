package api

import (
	"net/http"

	"uniwish.com/internal/api/handlers"
)

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", handlers.Health)
}
