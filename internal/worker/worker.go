/*
uniwish.com/interal/worker/worker

centralized scraper worker logic
*/
package worker

import (
	"context"
	"fmt"

	"uniwish.com/internal/api/repository"
	"uniwish.com/internal/scrapers"
)

type Worker struct {
	repo     WorkerRepo
	registry scrapers.Registry
}

func NewWorker(
	wr WorkerRepo,
	registry scrapers.Registry,
) *Worker {
	return &Worker{repo: wr, registry: registry}
}

func (w *Worker) RunOnce(ctx context.Context) error {
	job, err := w.ClaimJob(ctx)
	if err != nil {
		return fmt.Errorf("claim job error: %w", err)
	}

	err = w.ProcessJob(ctx, job)
	if err != nil {
		return fmt.Errorf("process job error: %w", err)
	}
	return nil
}

func (w *Worker) ClaimJob(ctx context.Context) (*repository.ScrapeRequest, error) {
	// create our repo and give us the transation it will use
	session, err := w.repo.BeginSession(ctx)

	if err != nil {
		return nil, err
	}

	// grab a job
	job, err := session.Dequeue(ctx)

	if err != nil {
		session.Rollback()
		return nil, fmt.Errorf("deqeue error: %w", err)
	}

	if job == nil {
		// though technically no need for rollback as nothing was updated
		// rollbacking just in case Dequeue changes before this does
		session.Rollback()
		return nil, ErrNoWork
	}

	if err = session.Commit(); err != nil {
		session.Rollback()
		return nil, fmt.Errorf("dequeue commit error: %w", err)
	}
	return job, nil
}

func (w *Worker) ProcessJob(ctx context.Context, job *repository.ScrapeRequest) error {
	session, err := w.repo.BeginSession(ctx)

	if err != nil {
		return err
	}
	scraper, err := w.registry.NewScraperFor(job.URL)
	if err != nil {
		// dead letter and surpress unsupported urls
		session.MarkFailed(ctx, job.ID)
		session.Commit()
		return JobError{JobID: job.ID, Err: err, Kind: JobUnsupportedStore}
	}
	productRecord, err := scraper.Scrape(ctx, job.URL)
	if err != nil {
		// dead letter failing scrapes but escalate for logging
		// TODO consider different error codes to easily diagnose different fail cases
		session.MarkFailed(ctx, job.ID)
		session.Commit()
		return JobError{JobID: job.ID, Err: err, Kind: JobScrapeFailed}
	}

	// TODO add processing lease metadata on jobs for reaper processes
	// otherwise failure (transcient or not) on these tasks can leave zombie (eternally processing) jobs
	_, err = session.UpsertProduct(ctx, *productRecord.Product)
	if err != nil {
		session.Rollback()
		return fmt.Errorf("process job error: %w", err)
	}

	err = session.InsertPrice(ctx, *productRecord.Offers)

	if err != nil {
		session.Rollback()
		return fmt.Errorf("process job error: %w", err)
	}

	session.MarkDone(ctx, job.ID)
	if err = session.Commit(); err != nil {
		session.Rollback()
		return fmt.Errorf("process commit error: %w", err)
	}
	return nil
}
