/*
uniwish.com/interal/worker/worker

centralized scraper worker logic
*/
package worker

import (
	"context"
	"errors"
	"fmt"

	"uniwish.com/internal/api/repository"
	"uniwish.com/internal/api/services"
)

var ErrNoJob = errors.New("No Job available")

type Worker struct {
	repoWithTxFactory func() (repository.ScrapeRequestRepository, repository.Transaction, error)
	scraperFactory    func(string) (services.BaseScraper, error)
}

func NewWorker(
	repoWithTxFactory func() (repository.ScrapeRequestRepository, repository.Transaction, error),
	scraperFactory func(string) (services.BaseScraper, error),
) *Worker {
	return &Worker{repoWithTxFactory, scraperFactory}
}
func (w *Worker) ClaimJob(ctx context.Context) (*repository.ScrapeRequest, error) {
	repo, tx, err := w.repoWithTxFactory()

	if err != nil {
		return nil, err
	}

	job, err := repo.Dequeue(ctx)

	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("deqeue error %w", err)
	}
	if job == nil {
		tx.Rollback()
		return nil, ErrNoJob
	}
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("commit error %w", err)
	}
	return job, nil
}

func (w *Worker) ProcessJob(ctx context.Context, job *repository.ScrapeRequest) error {
	repo, tx, err := w.repoWithTxFactory()

	if err != nil {
		return err
	}
	scraper, err := w.scraperFactory(job.URL)
	if err != nil {
		repo.MarkFailed(ctx, job.ID)
		tx.Commit()
		return fmt.Errorf("scraper factory error %w", err)
	}

	_, err = scraper.Scrape(ctx, job.URL)
	if err != nil {
		repo.MarkFailed(ctx, job.ID)
		tx.Commit()
		return fmt.Errorf("scrape error %w", err)
	}

	repo.MarkDone(ctx, job.ID)
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("commit error %w", err)
	}
	return nil
}

func (w *Worker) RunOnce(ctx context.Context) error {
	job, err := w.ClaimJob(ctx)
	if err != nil {
		return fmt.Errorf("claim job error %w", err)
	}
	err = w.ProcessJob(ctx, job)
	if err != nil {
		return fmt.Errorf("process job error %w", err)
	}
	return nil
}
