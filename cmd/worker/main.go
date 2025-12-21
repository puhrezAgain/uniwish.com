/*
uniwish.com/interal/cmd/worker/main

entrypoint to kick off scraper worker
*/
package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"uniwish.com/internal/api/config"
	"uniwish.com/internal/api/repository"
)

func main() {
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

	errCh := make(chan error, 1)
	go func() {
		repo := repository.NewPostgresScrapeRequestRepository(db)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				job, err := repo.Dequeue(ctx)
				if err != nil {
					logger.Error(
						"worker faces job error, making failed",
						"error", err,
						"id", job.ID)
					repo.MarkFailed(ctx, job.ID)
					continue
				}
				if job == nil {
					time.Sleep(cfg.WORKER_POLL_INTERVAL)
					continue
				}
				// TODO: scraper = scraperregistry.get(job.url.host)
				// scraper.scrape(job.url)
				// repo.MakeDone(ctx, job.ID)
				continue
			}
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		logger.Error("worker error", "err", err)
	}
}
