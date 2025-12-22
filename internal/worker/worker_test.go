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
	"uniwish.com/internal/api/services"
)

func TestRunOnce(t *testing.T) {
	tests := []struct {
		name                string
		repo                FakeRepo
		trans               FakeTransaction
		scraper             FakeScraper
		factoryError        error
		expectedErr         error
		repoCalls           []string
		transCalls          []string
		scraperFactoryCalls []string
		scraperCalls        []string
	}{
		{
			name:         "success",
			repo:         &DefaultFakeRepo{},
			trans:        &DefaultFakeTransaction{},
			scraper:      &DefaultFakeScraper{},
			factoryError: nil,
			expectedErr:  nil,
			repoCalls: []string{
				"Dequeue",
				"UpsertProduct",
				"InsertPrice",
				"MarkDone",
			},
			transCalls: []string{
				"Commit",
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
			trans:        &DefaultFakeTransaction{},
			scraper:      &DefaultFakeScraper{},
			factoryError: nil,
			expectedErr:  ErrNoWork,
			repoCalls: []string{
				"Dequeue",
			},
			transCalls: []string{
				"Rollback",
			},
			scraperFactoryCalls: []string{},
			scraperCalls:        []string{},
		},
		{
			name:         "not_supported",
			repo:         &DefaultFakeRepo{},
			trans:        &DefaultFakeTransaction{},
			scraper:      &DefaultFakeScraper{},
			factoryError: errors.ErrStoreUnsupported,
			expectedErr:  errors.ErrStoreUnsupported,
			repoCalls: []string{
				"Dequeue",
				"MarkFailed",
			},
			transCalls: []string{
				"Commit",
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
			trans:        &DefaultFakeTransaction{},
			scraper:      &DefaultFakeScraper{},
			factoryError: nil,
			expectedErr:  sql.ErrConnDone,
			repoCalls: []string{
				"Dequeue",
			},
			transCalls: []string{
				"Rollback",
			},
			scraperFactoryCalls: []string{},
			scraperCalls:        []string{},
		},
		{
			name:         "faulty_commit",
			repo:         &DefaultFakeRepo{},
			trans:        &FaultyCommitTransaction{},
			scraper:      &DefaultFakeScraper{},
			factoryError: nil,
			expectedErr:  sql.ErrTxDone,
			repoCalls: []string{
				"Dequeue",
			},
			transCalls: []string{
				"Commit",
				"Rollback",
			},
			scraperFactoryCalls: []string{},
			scraperCalls:        []string{},
		},
		{
			name:         "faulty_scrape",
			repo:         &DefaultFakeRepo{},
			trans:        &DefaultFakeTransaction{},
			scraper:      &FakeFaultyScraper{},
			factoryError: nil,
			expectedErr:  errors.ErrScrapeFailed,
			repoCalls: []string{
				"Dequeue",
				"MarkFailed",
			},
			transCalls: []string{
				"Commit",
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
			repoFactory := NewFakeRepoFactory(tt.repo, tt.trans)
			scraperFactory := NewFakeScraperFactory(tt.scraper, tt.factoryError)
			worker := Worker{repoFactory, scraperFactory.scaperFactoryFunc}
			rv := worker.RunOnce(context.Background())
			if !goErrors.Is(rv, tt.expectedErr) {
				t.Fatalf("expected rv %v received %v", tt.expectedErr, rv)
			}
			if !slices.Equal(tt.repoCalls, tt.repo.Calls()) {
				t.Fatalf("Expected %q, received %q", tt.repoCalls, tt.repo.Calls())
			}
			if !slices.Equal(tt.transCalls, tt.trans.Calls()) {
				t.Fatalf("Expected %q, received %q", tt.transCalls, tt.trans.Calls())
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
	repo := &DefaultFakeRepo{}
	trans := &FaultyCommitTransaction{}
	scraper := &DefaultFakeScraper{}
	repoFactory := NewFakeRepoFactory(repo, trans)
	scraperFactory := NewFakeScraperFactory(scraper, nil)
	worker := Worker{repoFactory, scraperFactory.scaperFactoryFunc}

	worker.ProcessJob(context.Background(), NewFakeJob())
	if !slices.Contains(repo.Calls(), "MarkDone") {
		t.Fatal("markdone not called")
	}
	if !slices.Contains(trans.Calls(), "Commit") {
		t.Fatal("trans commit not called")
	}
	if !slices.Contains(trans.Calls(), "Rollback") {
		t.Fatal("rollback not called")
	}
}

func TestClaimJob_FaultyTx(t *testing.T) {
	repo := &DefaultFakeRepo{}
	trans := &DefaultFakeTransaction{}
	repoFactoryWithError := func() (WorkerRepository, repository.Transaction, error) {
		return repo, trans, sql.ErrConnDone
	}
	stubScraperFactory := func(_ string) (services.BaseScraper, error) {
		return nil, nil
	}
	worker := Worker{repoFactoryWithError, stubScraperFactory}

	_, err := worker.ClaimJob(context.Background())
	if err == nil {
		t.Fatal("error nil")
	}
	if slices.Contains(repo.Calls(), "Dequeue") {
		t.Fatal("dequeue called")
	}
}

func TestProcessJob_FaultyTx(t *testing.T) {
	scraperFactory := NewFakeScraperFactory(&DefaultFakeScraper{}, nil)
	stubRepoFactory := func() (WorkerRepository, repository.Transaction, error) {
		return nil, nil, sql.ErrTxDone
	}
	worker := Worker{stubRepoFactory, scraperFactory.scaperFactoryFunc}
	err := worker.ProcessJob(context.Background(), NewFakeJob())
	if err == nil {
		t.Fatal("error nil")
	}
	if slices.Contains(scraperFactory.Calls(), "New") {
		t.Fatal("factory called")
	}
}
