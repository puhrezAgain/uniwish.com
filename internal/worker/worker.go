/*
uniwish.com/interal/worker/worker

centralized scraper worker logic
*/
package worker

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"uniwish.com/internal/api/repository"
	"uniwish.com/internal/api/services"
)

var ErrIdle = errors.New("no Job available")

type JobError struct {
	// used to represent errors that shouldn't be considered worker critical
	// It indicates the worker is healthy and should continue running.
	JobID uuid.UUID
	Err   error
}

func (e JobError) Error() string {
	return fmt.Sprintf("job %s: %v", e.JobID, e.Err)
}

func (e JobError) Unwrap() error {
	return e.Err
}

type Worker struct {
	// in order to make our worker easily testible
	// we delegate repo and scraper configuration to instantiators
	repoWithTxFactory func() (repository.ScrapeRequestRepository, repository.Transaction, error)
	scraperFactory    func(string) (services.BaseScraper, error)
}

func NewWorker(
	repoWithTxFactory func() (repository.ScrapeRequestRepository, repository.Transaction, error),
	scraperFactory func(string) (services.BaseScraper, error),
) *Worker {
	return &Worker{repoWithTxFactory, scraperFactory}
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

func (w *Worker) ClaimJob(ctx context.Context) (*repository.ScrapeRequest, error) {
	// create our repo and give us the transation it will use
	repo, tx, err := w.repoWithTxFactory()

	if err != nil {
		return nil, err
	}

	// grab a job
	job, err := repo.Dequeue(ctx)

	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("deqeue error %w", err)
	}

	if job == nil {
		// no job no problem, but lets indicate that just in case there is problem
		tx.Rollback()
		return nil, ErrIdle
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
		// dead letter and surpress unsupported urls
		repo.MarkFailed(ctx, job.ID)
		tx.Commit()
		return JobError{JobID: job.ID, Err: err}
	}

	_, err = scraper.Scrape(ctx, job.URL)
	if err != nil {
		// dead letter failing scrapes but escalate for logging
		repo.MarkFailed(ctx, job.ID)
		tx.Commit()
		return JobError{JobID: job.ID, Err: err}
	}

	repo.MarkDone(ctx, job.ID)
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("commit error %w", err)
	}
	return nil
}
