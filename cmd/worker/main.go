/*
uniwish.com/interal/cmd/worker/main

entrypoint to kick off scraper worker
*/
package main

import (
	"context"
	"database/sql"
	goErrors "errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"uniwish.com/internal/api/config"
	"uniwish.com/internal/api/errors"
	"uniwish.com/internal/api/repository"
	"uniwish.com/internal/api/services"
	"uniwish.com/internal/worker"
)

func main() {
	/*
		sets up logger, config, and database
		before creating the worker which runs in a loop in a goroutine
		gracefully shutdowns according to context
	*/
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()

	if err != nil {
		logger.Error("configuration load fail", "err", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DBURL)

	if err != nil {
		logger.Error("db open failed", "err", err)
		os.Exit(1)
	}

	defer db.Close()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	if err := db.PingContext(ctx); err != nil {
		logger.Error("db ping failed", "err", err)
		os.Exit(1)
	}

	// our worker require us to dynamically create our repo
	// in order to simplfy testability
	// we use it here to ensure all work goes into transactions
	repoWithTxFactory := func() (repository.ScrapeRequestRepository, repository.Transaction, error) {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return nil, nil, err
		}

		return repository.NewPostgresScrapeRequestRepository(tx), tx, nil
	}
	go func() {
		worker := worker.NewWorker(repoWithTxFactory, services.NewScraper)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				err = worker.RunOnce(ctx)
				if goErrors.Is(err, errors.ErrNoJob) {
					time.Sleep(cfg.WorkerPollInterval) // TODO: backoff on repeated errors
					continue
				}
				if err != nil {
					logger.Error("worker error", "error", err)
					time.Sleep(cfg.WorkerPollInterval) // TODO: backoff on repeated errors
				}
			}
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")
}
