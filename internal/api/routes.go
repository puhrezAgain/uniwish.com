/*
uniwish.com/interal/api/routes

centralizes logic concerned with routing
*/
package api

import (
	"database/sql"
	"net/http"

	"uniwish.com/internal/api/handlers"
	"uniwish.com/internal/api/repository"
	"uniwish.com/internal/api/services"
)

func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	healthServce := services.NewHealthService()
	healthHandler := handlers.NewHealthHandler(healthServce)
	mux.Handle("/health", healthHandler)

	repo := repository.NewPostgresScrapeRequestRepository(db)
	scrapeRequestService := services.NewScrapeRequestService(repo)
	scrapeRequestHandler := handlers.NewCreateItemHandler(scrapeRequestService)
	mux.Handle("/scrape-requests", scrapeRequestHandler)
}
