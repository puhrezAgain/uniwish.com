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
	"uniwish.com/internal/domain"
	"uniwish.com/internal/scrapers"
)

func NewFakeJob() *repository.ScrapeRequest {
	return &repository.ScrapeRequest{
		ID:     uuid.New(),
		URL:    "http://store.com",
		Status: "pending",
	}
}

func NewFakeProduct() *domain.ProductRecord {
	product := &domain.ProductSnapshot{
		ID:       uuid.New(),
		Store:    "zara",
		SKU:      "12345",
		Name:     "Zara jacket",
		ImageURL: "http://zara.com/img.jpeg",
	}

	offers := &[]domain.Offer{
		{
			Price:    45.32,
			Currency: "EUR",
		},
		{
			Price:    25.32,
			Currency: "EUR",
		},
	}
	return &domain.ProductRecord{Product: product, Offers: offers}
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

type FakeRepo interface {
	WorkerRepo
	Session() FakeWorkerSession
}

type DefaultFakeRepo struct {
	DefaultWorkerRepo
	session FakeWorkerSession
}

func (wr *DefaultFakeRepo) Session() FakeWorkerSession {
	return wr.session
}

func (wr *DefaultFakeRepo) BeginSession(ctx context.Context) (WorkerSession, error) {
	if wr.session == nil {
		wr.session = &DefaultFakeWorkerSession{}
	}
	return wr.session, nil
}

type FakeWorkerSession interface {
	repository.ScrapeRequestRepository
	repository.ProductRepository
	repository.Transaction
	Calls() []string
}
type DefaultFakeWorkerSession struct {
	CallRecorder
}

func (r *DefaultFakeWorkerSession) Insert(ctx context.Context, url string) (uuid.UUID, error) {
	r.record("Insert")
	return uuid.New(), nil
}
func (r *DefaultFakeWorkerSession) Dequeue(ctx context.Context) (*repository.ScrapeRequest, error) {
	r.record("Dequeue")
	return NewFakeJob(), nil
}
func (r *DefaultFakeWorkerSession) MarkDone(ctx context.Context, id uuid.UUID) error {
	r.record("MarkDone")
	return nil
}
func (r *DefaultFakeWorkerSession) MarkFailed(ctx context.Context, id uuid.UUID) error {
	r.record("MarkFailed")
	return nil
}

func (r *DefaultFakeWorkerSession) UpsertProduct(context.Context, domain.ProductSnapshot) (uuid.UUID, error) {
	r.record("UpsertProduct")
	return uuid.Nil, nil
}
func (r *DefaultFakeWorkerSession) InsertPrice(context.Context, domain.Offer) error {
	r.record("InsertPrice")
	return nil
}

func (t *DefaultFakeWorkerSession) Rollback() error {
	t.record("Rollback")
	return nil
}
func (t *DefaultFakeWorkerSession) Commit() error {
	t.record("Commit")
	return nil
}

type FakeNoJobsRepo struct {
	DefaultFakeRepo
}

func (wr *FakeNoJobsRepo) BeginSession(ctx context.Context) (WorkerSession, error) {
	if wr.session == nil {
		wr.session = &FakeNoJobsRepoSession{}
	}
	return wr.session, nil
}

type FakeNoJobsRepoSession struct {
	DefaultFakeWorkerSession
}

func (f *FakeNoJobsRepoSession) Dequeue(ctx context.Context) (*repository.ScrapeRequest, error) {
	f.DefaultFakeWorkerSession.Dequeue(ctx)
	return nil, nil
}

type FakeDBErrorRepo struct {
	DefaultFakeRepo
}

func (wr *FakeDBErrorRepo) BeginSession(ctx context.Context) (WorkerSession, error) {
	if wr.session == nil {
		wr.session = &FakeDBErrorRepoSession{}
	}
	return wr.session, nil
}

type FakeDBErrorRepoSession struct {
	DefaultFakeWorkerSession
}

func (f *FakeDBErrorRepoSession) Dequeue(ctx context.Context) (*repository.ScrapeRequest, error) {
	f.DefaultFakeWorkerSession.Dequeue(ctx)
	return nil, sql.ErrConnDone
}

type FaultyTransactionRepo struct {
	DefaultFakeRepo
}

func (f *FaultyTransactionRepo) BeginSession(ctx context.Context) (WorkerSession, error) {
	return nil, sql.ErrConnDone
}

type FaultyCommitRepo struct {
	DefaultFakeRepo
}

func (wr *FaultyCommitRepo) BeginSession(ctx context.Context) (WorkerSession, error) {
	if wr.session == nil {
		wr.session = &FaultyCommitRepoSession{}
	}
	return wr.session, nil
}

type FaultyCommitRepoSession struct {
	DefaultFakeWorkerSession
}

func (t *FaultyCommitRepoSession) Commit() error {
	t.DefaultFakeWorkerSession.Commit()
	return sql.ErrTxDone
}

type FakeScraper interface {
	scrapers.Scraper
	Calls() []string
}

type DefaultFakeScraper struct {
	scrapers.DefaultScraper
	CallRecorder
}

func (s *DefaultFakeScraper) Scrape(ctx context.Context, url string) (*domain.ProductRecord, error) {
	s.record("Scrape")
	return NewFakeProduct(), nil
}

type FakeFaultyScraper struct {
	DefaultFakeScraper
}

func (s *FakeFaultyScraper) Scrape(ctx context.Context, url string) (*domain.ProductRecord, error) {
	s.DefaultFakeScraper.Scrape(ctx, url)
	return nil, errors.ErrScrapeFailed
}

type FakeScraperFactory struct {
	scraper           scrapers.Scraper
	scaperFactoryFunc func(string) (scrapers.Scraper, error)
	CallRecorder
}

func NewFakeScraperFactory(scraper scrapers.Scraper, factoryError error) *FakeScraperFactory {
	sf := FakeScraperFactory{
		scraper: scraper,
	}
	sf.scaperFactoryFunc = func(URL string) (scrapers.Scraper, error) {
		sf.calls = append(sf.calls, "New")

		return sf.scraper, factoryError
	}
	return &sf
}

func NewUnsupoportedScraper(_ string) (scrapers.Scraper, error) {
	return nil, errors.ErrStoreUnsupported
}
