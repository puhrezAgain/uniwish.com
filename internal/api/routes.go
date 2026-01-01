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
	"uniwish.com/internal/scrapers"
)

func RegisterRoutes(mux *http.ServeMux, db *sql.DB, registry scrapers.Registry) {
	healthServce := services.NewHealthService()
	healthHandler := handlers.NewHealthHandler(healthServce)
	mux.Handle("/health", healthHandler)

	scrapeRepo := repository.NewPostgresScrapeRequestRepository(db)

	scrapeRequestService := services.NewScrapeRequestService(scrapeRepo, registry)
	scrapeRequestHandler := handlers.NewCreateItemHandler(scrapeRequestService)
	mux.Handle("/scrape-requests", scrapeRequestHandler)

	productRepo := repository.NewDefaultProductReader(db)
	productService := services.NewDefaultProductReaderService(productRepo)
	productHandler := handlers.NewDefaultProductHandler(productService)
	mux.HandleFunc("/products", productHandler.ListProducts)
	mux.HandleFunc("/products/{id}", productHandler.GetProduct)

}
