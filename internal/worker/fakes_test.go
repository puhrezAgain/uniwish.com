/*
uniwish.com/interal/worker/fakes_test

stubs for worker tests
*/
package worker

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"uniwish.com/internal/api/errors"
	"uniwish.com/internal/api/repository"
	"uniwish.com/internal/api/services"
	"uniwish.com/internal/domain"
)

func NewFakeJob() *repository.ScrapeRequest {
	return &repository.ScrapeRequest{
		ID:     uuid.New(),
		URL:    "http://store.com",
		Status: "pending",
	}
}

func NewFakeProduct() *domain.ProductSnapshot {
	return &domain.ProductSnapshot{
		ID:             uuid.New(),
		Store:          "zara",
		StoreProductID: "12345",
		Name:           "Zara jacket",
		ImageURL:       "http://zara.com/img.jpeg",
		Price:          45.32,
		Currency:       "euro",
	}
}

type FakeRepo interface {
	WorkerRepository
	Calls() []string
}

type CallRecorder struct {
	calls []string
}

func (s *CallRecorder) Calls() []string {
	return s.calls
}

func (s *CallRecorder) record(call string) {
	s.calls = append(s.calls, call)
}

type DefaultFakeRepo struct {
	WorkerRepository
	CallRecorder
}

func (r *DefaultFakeRepo) Insert(ctx context.Context, url string) (uuid.UUID, error) {
	r.record("Insert")
	return uuid.New(), nil
}
func (r *DefaultFakeRepo) Dequeue(ctx context.Context) (*repository.ScrapeRequest, error) {
	r.record("Dequeue")
	return NewFakeJob(), nil
}
func (r *DefaultFakeRepo) MarkDone(ctx context.Context, id uuid.UUID) error {
	r.record("MarkDone")
	return nil
}
func (r *DefaultFakeRepo) MarkFailed(ctx context.Context, id uuid.UUID) error {
	r.record("MarkFailed")
	return nil
}

func (r *DefaultFakeRepo) UpsertProduct(context.Context, domain.ProductSnapshot) (uuid.UUID, error) {
	r.record("UpsertProduct")
	return uuid.Nil, nil
}
func (r *DefaultFakeRepo) InsertPrice(context.Context, uuid.UUID, float32, string) error {
	r.record("InsertPrice")
	return nil
}

type FakeNoJobsRepo struct {
	DefaultFakeRepo
}

func (f *FakeNoJobsRepo) Dequeue(ctx context.Context) (*repository.ScrapeRequest, error) {
	f.DefaultFakeRepo.Dequeue(ctx)
	return nil, nil
}

type FakeDBErrorRepo struct {
	DefaultFakeRepo
}

func (f *FakeDBErrorRepo) Dequeue(ctx context.Context) (*repository.ScrapeRequest, error) {
	f.DefaultFakeRepo.Dequeue(ctx)
	return nil, sql.ErrConnDone
}

type FakeTransaction interface {
	repository.Transaction
	Calls() []string
}
type DefaultFakeTransaction struct {
	CallRecorder
}

func (t *DefaultFakeTransaction) Rollback() error {
	t.record("Rollback")
	return nil
}
func (t *DefaultFakeTransaction) Commit() error {
	t.record("Commit")
	return nil
}

type FaultyCommitTransaction struct {
	DefaultFakeTransaction
}

func (t *FaultyCommitTransaction) Commit() error {
	t.DefaultFakeTransaction.Commit()
	return sql.ErrTxDone
}

func NewFakeRepoFactory(repo WorkerRepository, trans repository.Transaction) func() (WorkerRepository, repository.Transaction, error) {
	return func() (WorkerRepository, repository.Transaction, error) {
		return repo, trans, nil
	}
}

type FakeScraper interface {
	services.BaseScraper
	Calls() []string
}

type DefaultFakeScraper struct {
	services.Scraper
	CallRecorder
}

func (s *DefaultFakeScraper) Scrape(ctx context.Context, url string) (*domain.ProductSnapshot, error) {
	s.record("Scrape")
	return NewFakeProduct(), nil
}

type FakeFaultyScraper struct {
	DefaultFakeScraper
}

func (s *FakeFaultyScraper) Scrape(ctx context.Context, url string) (*domain.ProductSnapshot, error) {
	s.DefaultFakeScraper.Scrape(ctx, url)
	return nil, errors.ErrScrapeFailed
}

type FakeScraperFactory struct {
	scraper           services.BaseScraper
	scaperFactoryFunc func(string) (services.BaseScraper, error)
	CallRecorder
}

func NewFakeScraperFactory(scraper services.BaseScraper, factoryError error) *FakeScraperFactory {
	sf := FakeScraperFactory{
		scraper: scraper,
	}
	sf.scaperFactoryFunc = func(URL string) (services.BaseScraper, error) {
		sf.calls = append(sf.calls, "New")

		return sf.scraper, factoryError
	}
	return &sf
}

func NewUnsupoportedScraper(_ string) (services.BaseScraper, error) {
	return nil, errors.ErrStoreUnsupported
}
