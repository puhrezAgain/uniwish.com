/*
uniwish.com/interal/api/routes

centralizes logic concerned with routing
*/
package api

import (
	"net/http"

	"uniwish.com/internal/api/handlers"
	"uniwish.com/internal/api/services"
)

func RegisterRoutes(mux *http.ServeMux) {
	healthServce := services.NewHealthService()
	healthHandler := handlers.NewHealthHandler(healthServce)
	mux.Handle("/health", healthHandler)

	itemService := services.NewScrapeRequestService()
	itemHandler := handlers.NewCreateItemHandler(itemService)
	mux.Handle("/items", itemHandler)
}
