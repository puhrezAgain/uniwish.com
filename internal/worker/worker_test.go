/*
uniwish.com/interal/worker/worker_test

tests for scraper worker logic
*/
package worker

import (
	"context"
	"database/sql"
	goErrors "errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"uniwish.com/internal/api/errors"
	"uniwish.com/internal/api/repository"
	"uniwish.com/internal/api/services"
)

type FakeRepo struct {
	InsertCalled     bool
	DequeueCalled    bool
	MarkDoneCalled   bool
	MarkFailedCalled bool
}

func NewFakeJob() *repository.ScrapeRequest {
	return &repository.ScrapeRequest{
		ID:     uuid.New(),
		URL:    "http://store.com",
		Status: "pending",
	}
}
func (r *FakeRepo) Insert(ctx context.Context, url string) (uuid.UUID, error) {
	r.InsertCalled = true
	return uuid.New(), nil
}
func (r *FakeRepo) Dequeue(ctx context.Context) (*repository.ScrapeRequest, error) {
	r.DequeueCalled = true
	return NewFakeJob(), nil
}
func (r *FakeRepo) MarkDone(ctx context.Context, id uuid.UUID) error {
	r.MarkDoneCalled = true
	return nil
}
func (r *FakeRepo) MarkFailed(ctx context.Context, id uuid.UUID) error {
	r.MarkFailedCalled = true
	return nil
}

type FakeTransaction struct {
	RollbackCalled bool
	CommitCalled   bool
}

func (t *FakeTransaction) Rollback() error {
	t.RollbackCalled = true
	return nil
}
func (t *FakeTransaction) Commit() error {
	t.CommitCalled = true
	return nil
}

func NewFakeRepoFactory(repo repository.ScrapeRequestRepository, trans repository.Transaction) func() (repository.ScrapeRequestRepository, repository.Transaction, error) {
	return func() (repository.ScrapeRequestRepository, repository.Transaction, error) {
		return repo, trans, nil
	}
}

type FakeScraper struct {
	ScrapeCalled bool
}

func (s *FakeScraper) Scrape(ctx context.Context, url string) (*services.PlaceholderProduct, error) {
	s.ScrapeCalled = true
	return nil, nil
}

type FakeScraperFactory struct {
	scraper           services.BaseScraper
	scaperFactoryFunc func(string) (services.BaseScraper, error)
	factoryCalled     bool
}

func NewFakeScraperFactory(scraper services.BaseScraper, factoryError error) *FakeScraperFactory {
	sf := FakeScraperFactory{
		scraper: scraper,
	}
	sf.scaperFactoryFunc = func(URL string) (services.BaseScraper, error) {
		sf.factoryCalled = true

		return sf.scraper, factoryError
	}
	return &sf
}

type FakeNoJobsRepo struct {
	FakeRepo
}

func (f *FakeNoJobsRepo) Dequeue(ctx context.Context) (*repository.ScrapeRequest, error) {
	f.DequeueCalled = true
	return nil, nil
}

type FakeDBErrorRepo struct {
	FakeRepo
}

func (f *FakeDBErrorRepo) Dequeue(ctx context.Context) (*repository.ScrapeRequest, error) {
	f.DequeueCalled = true
	return nil, sql.ErrConnDone
}

func NewUnsupoportedScraper(_ string) (services.BaseScraper, error) {
	return nil, errors.ErrStoreUnsupported
}

type FaultyCommitTransaction struct {
	FakeTransaction
}

func (t *FaultyCommitTransaction) Commit() error {
	t.CommitCalled = true
	return sql.ErrTxDone
}

type FakeFaultyScraper struct {
	ScrapeCalled bool
}

func (s *FakeFaultyScraper) Scrape(ctx context.Context, url string) (*services.PlaceholderProduct, error) {
	s.ScrapeCalled = true
	return nil, errors.ErrScrapeFailed
}

func getBoolField(v any, name string) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	field := rv.FieldByName(name)

	return field.Bool()

}
func TestRunOnce(t *testing.T) {
	type StringBools struct {
		Key   string
		Value bool
	}
	tests := []struct {
		name                             string
		repo                             repository.ScrapeRequestRepository
		trans                            repository.Transaction
		scraper                          services.BaseScraper
		factoryError                     error
		rv                               error
		repoKeyToExpectedValue           []StringBools
		transKeyToExpectedValue          []StringBools
		scraperFactoryKeyToExpectedValue []StringBools
		scraperKeyToExpectedValue        []StringBools
	}{
		{
			name:         "success",
			repo:         &FakeRepo{},
			trans:        &FakeTransaction{},
			scraper:      &FakeScraper{},
			factoryError: nil,
			rv:           nil,
			repoKeyToExpectedValue: []StringBools{
				{"DequeueCalled", true},
				{"MarkDoneCalled", true},
				{"MarkFailedCalled", false},
			},
			transKeyToExpectedValue: []StringBools{
				{"CommitCalled", true},
				{"RollbackCalled", false},
			},
			scraperFactoryKeyToExpectedValue: []StringBools{
				{"factoryCalled", true},
			},
			scraperKeyToExpectedValue: []StringBools{
				{"ScrapeCalled", true},
			},
		},
		{
			name:         "no_jobs",
			repo:         &FakeNoJobsRepo{},
			trans:        &FakeTransaction{},
			scraper:      &FakeScraper{},
			factoryError: nil,
			rv:           ErrIdle,
			repoKeyToExpectedValue: []StringBools{
				{"DequeueCalled", true},
				{"MarkDoneCalled", false},
				{"MarkFailedCalled", false},
			},
			transKeyToExpectedValue: []StringBools{
				{"CommitCalled", false},
				{"RollbackCalled", true},
			},
			scraperFactoryKeyToExpectedValue: []StringBools{
				{"factoryCalled", false},
			},
			scraperKeyToExpectedValue: []StringBools{
				{"ScrapeCalled", false},
			},
		},
		{
			name:         "not_supported",
			repo:         &FakeRepo{},
			trans:        &FakeTransaction{},
			scraper:      &FakeScraper{},
			factoryError: errors.ErrStoreUnsupported,
			rv:           errors.ErrStoreUnsupported,
			repoKeyToExpectedValue: []StringBools{
				{"DequeueCalled", true},
				{"MarkDoneCalled", false},
				{"MarkFailedCalled", true},
			},
			transKeyToExpectedValue: []StringBools{
				{"CommitCalled", true},
				{"RollbackCalled", false},
			},
			scraperFactoryKeyToExpectedValue: []StringBools{
				{"factoryCalled", true},
			},
			scraperKeyToExpectedValue: []StringBools{
				{"ScrapeCalled", false},
			},
		},
		{
			name:         "db_error",
			repo:         &FakeDBErrorRepo{},
			trans:        &FakeTransaction{},
			scraper:      &FakeScraper{},
			factoryError: nil,
			rv:           sql.ErrConnDone,
			repoKeyToExpectedValue: []StringBools{
				{"DequeueCalled", true},
				{"MarkDoneCalled", false},
				{"MarkFailedCalled", false},
			},
			transKeyToExpectedValue: []StringBools{
				{"CommitCalled", false},
				{"RollbackCalled", true},
			},
			scraperFactoryKeyToExpectedValue: []StringBools{
				{"factoryCalled", false},
			},
			scraperKeyToExpectedValue: []StringBools{
				{"ScrapeCalled", false},
			},
		},
		{
			name:         "faulty_commit",
			repo:         &FakeRepo{},
			trans:        &FaultyCommitTransaction{},
			scraper:      &FakeScraper{},
			factoryError: nil,
			rv:           sql.ErrTxDone,
			repoKeyToExpectedValue: []StringBools{
				{"DequeueCalled", true},
				{"MarkDoneCalled", false},
				{"MarkFailedCalled", false},
			},
			transKeyToExpectedValue: []StringBools{
				{"CommitCalled", true},
				{"RollbackCalled", true},
			},
			scraperFactoryKeyToExpectedValue: []StringBools{
				{"factoryCalled", false},
			},
			scraperKeyToExpectedValue: []StringBools{
				{"ScrapeCalled", false},
			},
		},
		{
			name:         "faulty_scrape",
			repo:         &FakeRepo{},
			trans:        &FakeTransaction{},
			scraper:      &FakeFaultyScraper{},
			factoryError: nil,
			rv:           errors.ErrScrapeFailed,
			repoKeyToExpectedValue: []StringBools{
				{"DequeueCalled", true},
				{"MarkDoneCalled", false},
				{"MarkFailedCalled", true},
			},
			transKeyToExpectedValue: []StringBools{
				{"CommitCalled", true},
				{"RollbackCalled", false},
			},
			scraperFactoryKeyToExpectedValue: []StringBools{
				{"factoryCalled", true},
			},
			scraperKeyToExpectedValue: []StringBools{
				{"ScrapeCalled", true},
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
			if !goErrors.Is(rv, tt.rv) {
				t.Fatalf("expected rv %v received %v", tt.rv, rv)
			}
			for _, p := range tt.repoKeyToExpectedValue {
				if getBoolField(tt.repo, p.Key) != p.Value {
					t.Fatalf("repo.%s is not %t", p.Key, p.Value)
				}
			}
			for _, p := range tt.transKeyToExpectedValue {
				if getBoolField(tt.trans, p.Key) != p.Value {
					t.Fatalf("trans.%s is not %t", p.Key, p.Value)
				}
			}
			for _, p := range tt.scraperFactoryKeyToExpectedValue {
				if getBoolField(scraperFactory, p.Key) != p.Value {
					t.Fatalf("scrapeFactory.%s is not %t", p.Key, p.Value)
				}
			}
			for _, p := range tt.scraperKeyToExpectedValue {
				if getBoolField(tt.scraper, p.Key) != p.Value {
					t.Fatalf("scraper.%s is not %t", p.Key, p.Value)
				}
			}
		})
	}
}

func TestProcessJob_FaultyCommit(t *testing.T) {
	repo := &FakeRepo{}
	trans := &FaultyCommitTransaction{}
	scraper := &FakeScraper{}
	repoFactory := NewFakeRepoFactory(repo, trans)
	scraperFactory := NewFakeScraperFactory(scraper, nil)
	worker := Worker{repoFactory, scraperFactory.scaperFactoryFunc}

	worker.ProcessJob(context.Background(), NewFakeJob())
	if !repo.MarkDoneCalled {
		t.Fatal("markdone not called")
	}
	if !trans.CommitCalled {
		t.Fatal("trans commit not called")
	}
	if !trans.RollbackCalled {
		t.Fatal("rollback not called")
	}
}

func TestClaimJob_FaultyTx(t *testing.T) {
	repo := &FakeRepo{}
	trans := &FakeTransaction{}
	repoFactoryWithError := func() (repository.ScrapeRequestRepository, repository.Transaction, error) {
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
	if repo.DequeueCalled {
		t.Fatal("dequeue called")
	}
}

func TestProcessJob_FaultyTx(t *testing.T) {
	scraperFactory := NewFakeScraperFactory(&FakeScraper{}, nil)
	stubRepoFactory := func() (repository.ScrapeRequestRepository, repository.Transaction, error) {
		return nil, nil, sql.ErrTxDone
	}
	worker := Worker{stubRepoFactory, scraperFactory.scaperFactoryFunc}
	err := worker.ProcessJob(context.Background(), NewFakeJob())
	if err == nil {
		t.Fatal("error nil")
	}
	if scraperFactory.factoryCalled {
		t.Fatal("factory called")
	}
}
