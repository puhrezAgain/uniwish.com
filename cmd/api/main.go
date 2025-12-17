package main

import (
	"log/slog"
	"net/http"
	"os"

	"uniwish.com/internal/api"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: api.NewServer(logger),
	}

	logger.Info("http server started", "addr", srv.Addr)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server unexpectedly closed", "err", err)
		os.Exit(1)
	}
}
