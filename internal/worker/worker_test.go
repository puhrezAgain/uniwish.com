/*
uniwish.com/interal/worker/worker_test

tests for scraper worker logic
*/
package worker

import (
	"context"
	"database/sql"
	goErrors "errors"
	"slices"
	"testing"

	"uniwish.com/internal/api/errors"
	"uniwish.com/internal/api/repository"
	"uniwish.com/internal/scrapers"
	"uniwish.com/internal/testutil"
)

func TestRunOnce(t *testing.T) {
	tests := []struct {
		name                string
		repo                FakeRepo
		scraper             FakeScraper
		factoryError        error
		expectedErr         error
		repoCalls           []string
		scraperFactoryCalls []string
		scraperCalls        []string
	}{
		{
			name:         "success",
			repo:         &DefaultFakeRepo{},
			scraper:      &DefaultFakeScraper{},
			factoryError: nil,
			expectedErr:  nil,
			repoCalls: []string{
				"Dequeue",
				"Commit",
				"UpsertProduct",
				"InsertPrice",
				"InsertPrice",
				"MarkDone",
				"Commit",
			},
			scraperFactoryCalls: []string{
				"New",
			},
			scraperCalls: []string{
				"Scrape",
			},
		},
		{
			name:         "no_jobs",
			repo:         &FakeNoJobsRepo{},
			scraper:      &DefaultFakeScraper{},
			factoryError: nil,
			expectedErr:  ErrNoWork,
			repoCalls: []string{
				"Dequeue",
				"Rollback",
			},
			scraperFactoryCalls: []string{},
			scraperCalls:        []string{},
		},
		{
			name:         "not_supported",
			repo:         &DefaultFakeRepo{},
			scraper:      &DefaultFakeScraper{},
			factoryError: errors.ErrStoreUnsupported,
			expectedErr:  errors.ErrStoreUnsupported,
			repoCalls: []string{
				"Dequeue",
				"Commit",
				"MarkFailed",
				"Commit",
			},
			scraperFactoryCalls: []string{
				"New",
			},
			scraperCalls: []string{},
		},
		{
			name:         "db_error",
			repo:         &FakeDBErrorRepo{},
			scraper:      &DefaultFakeScraper{},
			factoryError: nil,
			expectedErr:  sql.ErrConnDone,
			repoCalls: []string{
				"Dequeue",
				"Rollback",
			},
			scraperFactoryCalls: []string{},
			scraperCalls:        []string{},
		},
		{
			name:         "faulty_commit",
			repo:         &FaultyCommitRepo{},
			scraper:      &DefaultFakeScraper{},
			factoryError: nil,
			expectedErr:  sql.ErrTxDone,
			repoCalls: []string{
				"Dequeue",
				"Commit",
				"Rollback",
			},
			scraperFactoryCalls: []string{},
			scraperCalls:        []string{},
		},
		{
			name:         "faulty_scrape",
			repo:         &DefaultFakeRepo{},
			scraper:      &FakeFaultyScraper{},
			factoryError: nil,
			expectedErr:  errors.ErrScrapeFailed,
			repoCalls: []string{
				"Dequeue",
				"Commit",
				"MarkFailed",
				"Commit",
			},
			scraperFactoryCalls: []string{
				"New",
			},
			scraperCalls: []string{
				"Scrape",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			scraperFactory := NewFakeScraperFactory(tt.scraper, tt.factoryError)
			worker := NewWorker(tt.repo, scraperFactory.scaperFactoryFunc)
			rv := worker.RunOnce(context.Background())
			if !goErrors.Is(rv, tt.expectedErr) {
				t.Fatalf("expected rv %v received %v", tt.expectedErr, rv)
			}
			if !slices.Equal(tt.repoCalls, tt.repo.Session().Calls()) {
				t.Fatalf("Expected %q, received %q", tt.repoCalls, tt.repo.Session().Calls())
			}
			if !slices.Equal(tt.scraperCalls, tt.scraper.Calls()) {
				t.Fatalf("Expected %q, received %q", tt.scraperCalls, tt.scraper.Calls())
			}
			if !slices.Equal(tt.scraperFactoryCalls, scraperFactory.Calls()) {
				t.Fatalf("Expected %q, received %q", tt.scraperFactoryCalls, scraperFactory.Calls())
			}
		})
	}
}

func TestProcessJob_FaultyCommit(t *testing.T) {
	repo := &FaultyCommitRepo{}
	scraper := &DefaultFakeScraper{}
	scraperFactory := NewFakeScraperFactory(scraper, nil)
	worker := NewWorker(repo, scraperFactory.scaperFactoryFunc)
	expectedCalls := []string{"UpsertProduct", "InsertPrice", "InsertPrice", "MarkDone", "Commit", "Rollback"}
	worker.ProcessJob(context.Background(), NewFakeJob())
	if !slices.Equal(repo.Session().Calls(), expectedCalls) {
		t.Fatalf("Expected %q, received %q", expectedCalls, repo.Session().Calls())
	}
}

func TestClaimJob_FaultyTx(t *testing.T) {
	repo := &FaultyTransactionRepo{}
	stubScraperFactory := func(_ string) (scrapers.Scraper, error) {
		return nil, nil
	}
	worker := NewWorker(repo, stubScraperFactory)

	_, err := worker.ClaimJob(context.Background())
	if err == nil {
		t.Fatal("error nil")
	}
}

func TestProcessJob_FaultyTx(t *testing.T) {
	repo := &FaultyTransactionRepo{}
	scraperFactory := NewFakeScraperFactory(&DefaultFakeScraper{}, nil)
	worker := NewWorker(repo, scraperFactory.scaperFactoryFunc)
	err := worker.ProcessJob(context.Background(), NewFakeJob())
	if err == nil {
		t.Fatal("error nil")
	}
	if slices.Contains(scraperFactory.Calls(), "New") {
		t.Fatal("factory called")
	}
}

func TestRunOnce_Integration(t *testing.T) {
	testutil.RequireIntegration(t)
	testutil.TruncateTables(t, testDB)
	t.Cleanup(func() {
		testutil.TruncateTables(t, testDB)
	})

	repo := repository.NewPostgresScrapeRequestRepository(testDB)
	expectedUrl := "http://store.com"
	ctx := context.Background()

	id, err := repo.Insert(ctx, expectedUrl)

	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	scraper := &DefaultFakeScraper{}
	scraperFactory := NewFakeScraperFactory(scraper, nil)
	workerRepo := NewWorkerRepo(testDB)
	newWorker := NewWorker(workerRepo, scraperFactory.scaperFactoryFunc)

	newWorker.RunOnce(context.Background())

	var status string
	err = testDB.QueryRow(`
	SELECT status
	FROM scrape_requests
	WHERE id = $1
	`, id).Scan(&status)

	if status != "done" {
		t.Fatalf("expected status to be 'done', received, %s", status)
	}
}
