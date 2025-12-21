/*
uniwish.com/interal/worker/worker

centralized scraper worker logic
*/
package worker

import (
	"context"
	"errors"

	"uniwish.com/internal/api/repository"
	"uniwish.com/internal/api/services"
)

var ErrNoJob = errors.New("No Job available")

func ClaimJob(
	ctx context.Context,
	repoWithTxFactory func() (repository.ScrapeRequestRepository, repository.Transaction, error),
) (*repository.ScrapeRequest, error) {
	repo, tx, err := repoWithTxFactory()

	if err != nil {
		return nil, err
	}

	job, err := repo.Dequeue(ctx)

	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if job == nil {
		tx.Rollback()
		return nil, ErrNoJob
	}
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return nil, err
	}
	return job, nil
}

func ProcessJob(
	job *repository.ScrapeRequest, ctx context.Context,
	scraperFactory func(string) (services.BaseScraper, error),
	repoWithTxFactory func() (repository.ScrapeRequestRepository, repository.Transaction, error),
) error {
	repo, tx, err := repoWithTxFactory()

	if err != nil {
		return err
	}
	scraper, err := scraperFactory(job.URL)
	if err != nil {
		repo.MarkFailed(ctx, job.ID)
		return err
	}

	_, err = scraper.Scrape(ctx, job.URL)
	if err != nil {
		repo.MarkFailed(ctx, job.ID)
		return err
	}

	repo.MarkDone(ctx, job.ID)
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func RunOnce(ctx context.Context,
	scraperFactory func(string) (services.BaseScraper, error),
	repoWithTxFactory func() (repository.ScrapeRequestRepository, repository.Transaction, error),
) error {
	job, err := ClaimJob(ctx, repoWithTxFactory)
	if err != nil {
		return err
	}
	err = ProcessJob(job, ctx, scraperFactory, repoWithTxFactory)
	if err != nil {
		return err
	}
	return nil
}
