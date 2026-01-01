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
		name                   string
		repo                   FakeRepo
		scraper                FakeScraper
		scraperRegistryFactory func(FakeScraper) *scrapers.ScraperRegistry
		expectedErr            error
		repoCalls              []string
		scraperCalls           []string
	}{
		{
			name: "success",
			repo: &DefaultFakeRepo{},
			repoCalls: []string{
				"Dequeue",
				"Commit",
				"UpsertProduct",
				"InsertPrice",
				"MarkDone",
				"Commit",
			},
			scraperCalls: []string{
				"New",
				"Scrape",
			},
		},
		{
			name:        "no_jobs",
			repo:        &FakeNoJobsRepo{},
			expectedErr: ErrNoWork,
			repoCalls: []string{
				"Dequeue",
				"Rollback",
			},
			scraperCalls: []string{},
		},
		{
			name:                   "not_supported",
			repo:                   &DefaultFakeRepo{},
			scraperRegistryFactory: NewEmptyScraperRegistry,
			expectedErr:            scrapers.ErrNoScraper,
			repoCalls: []string{
				"Dequeue",
				"Commit",
				"MarkFailed",
				"Commit",
			},
			scraperCalls: []string{},
		},
		{
			name:        "db_error",
			repo:        &FakeDBErrorRepo{},
			expectedErr: sql.ErrConnDone,
			repoCalls: []string{
				"Dequeue",
				"Rollback",
			},
			scraperCalls: []string{},
		},
		{
			name:        "faulty_commit",
			repo:        &FaultyCommitRepo{},
			expectedErr: sql.ErrTxDone,
			repoCalls: []string{
				"Dequeue",
				"Commit",
				"Rollback",
			},
			scraperCalls: []string{},
		},
		{
			name:        "faulty_scrape",
			repo:        &DefaultFakeRepo{},
			scraper:     &FakeFaultyScraper{},
			expectedErr: errors.ErrScrapeFailed,
			repoCalls: []string{
				"Dequeue",
				"Commit",
				"MarkFailed",
				"Commit",
			},
			scraperCalls: []string{
				"New",
				"Scrape",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			scraper := tt.scraper
			if scraper == nil {
				scraper = &DefaultFakeScraper{}
			}
			f := tt.scraperRegistryFactory
			if f == nil {
				f = NewFakeScraperRegistry
			}

			registry := f(scraper)
			worker := NewWorker(tt.repo, registry)
			rv := worker.RunOnce(context.Background())
			if !goErrors.Is(rv, tt.expectedErr) {
				t.Fatalf("expected rv %v received %v", tt.expectedErr, rv)
			}
			if !slices.Equal(tt.repoCalls, tt.repo.Session().Calls()) {
				t.Fatalf("Expected %q, received %q", tt.repoCalls, tt.repo.Session().Calls())
			}
			if !slices.Equal(tt.scraperCalls, scraper.Calls()) {
				t.Fatalf("Expected %q, received %q", tt.scraperCalls, scraper.Calls())
			}
		})
	}
}

func TestProcessJob_FaultyCommit(t *testing.T) {
	repo := &FaultyCommitRepo{}
	scraper := &DefaultFakeScraper{}
	registry := NewFakeScraperRegistry(scraper)
	worker := NewWorker(repo, registry)

	expectedCalls := []string{"UpsertProduct", "InsertPrice", "MarkDone", "Commit", "Rollback"}
	worker.ProcessJob(context.Background(), NewFakeJob())
	if !slices.Equal(repo.Session().Calls(), expectedCalls) {
		t.Fatalf("Expected %q, received %q", expectedCalls, repo.Session().Calls())
	}
}

func TestClaimJob_FaultyTx(t *testing.T) {
	repo := &FaultyTransactionRepo{}
	scraper := &DefaultFakeScraper{}
	registry := NewFakeScraperRegistry(scraper)

	worker := NewWorker(repo, registry)

	_, err := worker.ClaimJob(context.Background())
	if err == nil {
		t.Fatal("error nil")
	}
}

func TestProcessJob_FaultyTx(t *testing.T) {
	repo := &FaultyTransactionRepo{}
	scraper := &DefaultFakeScraper{}
	registry := NewFakeScraperRegistry(scraper)
	worker := NewWorker(repo, registry)

	err := worker.ProcessJob(context.Background(), NewFakeJob())
	if err == nil {
		t.Fatal("error nil")
	}
	if slices.Contains(scraper.Calls(), "New") {
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
	registry := NewFakeScraperRegistry(scraper)
	workerRepo := NewWorkerRepo(testDB)
	newWorker := NewWorker(workerRepo, registry)

	err = newWorker.RunOnce(context.Background())

	if err != nil {
		t.Fatalf("expected nil err, recevied %v", err)
	}
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
