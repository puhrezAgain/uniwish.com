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
	"sync/atomic"
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
		scrapeWorker := worker.NewWorker(repoWithTxFactory, services.NewScraper)
		var failures atomic.Int32

		for {
			select {
			case <-ctx.Done():
				return
			default:
				err = scrapeWorker.RunOnce(ctx)
				var je worker.JobError
				switch {
				case err == nil:
					failures.Store(0)
				case goErrors.Is(err, errors.ErrIdle):
					// No work is not a failure nor a success
					// so lets not reset the failure counter, but also not increment it
				case goErrors.As(err, &je):
					// handling dead letter is proper health
					// resetting the failure counter because processing problems could be input issues
					logger.Error("job error", "error", err)
					failures.Store(0)
				default:
					logger.Error("worker error", "error", err)

					if failures.Add(1) >= int32(cfg.WorkerFailureTolerance) {
						logger.Error("worker tolerance exceeded", "failures", failures.Load())
						stop()
						return
					}
				}
				time.Sleep(cfg.WorkerPollInterval)
			}
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")
}
