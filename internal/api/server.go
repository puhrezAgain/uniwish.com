/*
uniwish.com/internal/api/server

Contains our wrapper of http.Server adding logging capacity, routes and middleware
*/
package api

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"

	"uniwish.com/internal/api/config"
	"uniwish.com/internal/api/middleware"
	"uniwish.com/internal/scrapers"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

func (s *Server) ListenAndServe() error {
	s.logger.Info("http server started", "addr", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(shutdownCtx context.Context) error {
	s.logger.Info("http server shutting down")
	return s.httpServer.Shutdown(shutdownCtx)
}
func NewServer(cfg *config.Config, logger *slog.Logger, db *sql.DB, registry scrapers.Registry) *Server {
	mux := http.NewServeMux()

	RegisterRoutes(mux, db, registry)

	handler := middleware.Logging(logger)(mux)

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Port),
		Handler: handler,
	}

	return &Server{
		httpServer: srv,
		logger:     logger,
	}
}
