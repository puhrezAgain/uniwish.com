/*
uniwish.com/interal/worker/supervisor_test

tests for worker supervisor
*/
package worker

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"testing"
	"time"
)

type FakeWorker struct {
	Results []error
	Calls   int
}

func (w *FakeWorker) RunOnce(ctx context.Context) error {
	if w.Calls >= len(w.Results) {
		return nil
	}

	err := w.Results[w.Calls]
	w.Calls++
	return err
}

type BlockingWorker struct{}

func (w *BlockingWorker) RunOnce(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
func TestWorkerSupervisor(t *testing.T) {
	tests := []struct {
		name   string
		worker FakeWorker
	}{
		{
			name: "after_threshold",
			worker: FakeWorker{
				Results: []error{
					sql.ErrConnDone,
					sql.ErrConnDone,
					sql.ErrConnDone,
				},
			},
		},
		{
			name: "jobError_resets_counter",
			worker: FakeWorker{
				Results: []error{
					sql.ErrConnDone,
					sql.ErrConnDone,
					JobError{},
					sql.ErrConnDone,
					sql.ErrConnDone,
					sql.ErrConnDone,
				},
			},
		},
		{
			name: "counter_doesnt_reset_on_idle",
			worker: FakeWorker{
				Results: []error{
					sql.ErrConnDone,
					sql.ErrConnDone,
					ErrIdle,
					sql.ErrConnDone,
				},
			},
		},
		{
			name: "counter_reset_on_success",
			worker: FakeWorker{
				Results: []error{
					sql.ErrConnDone,
					sql.ErrConnDone,
					nil,
					sql.ErrConnDone,
					sql.ErrConnDone,
					sql.ErrConnDone,
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		stopped := false
		ws := WorkerSupervisor{
			Worker:           &tt.worker,
			PollInterval:     0,
			FailureTolerance: 3,
			Sleep:            func(time.Duration) {},
			OnFatal:          func() { stopped = true },
			Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ws.Run(ctx)
		if !stopped {
			t.Fatal("expected supervisor to stop after 3 failures")
		}
	}
}

func TestWorkerSupervisor_StopsOnCancel(t *testing.T) {
	w := BlockingWorker{}
	ws := WorkerSupervisor{
		Worker:           &w,
		PollInterval:     0,
		FailureTolerance: 3,
		Sleep:            func(time.Duration) {},
		OnFatal:          func() {},
		Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	ctx, cancel := context.WithCancel(context.Background())

	go ws.Run(ctx)
	time.Sleep(20 * time.Millisecond)

	cancel()
	select {
	case <-ctx.Done():
		// success
	case <-time.After(200 * time.Millisecond):
		t.Fatal("supervisor did not end on context cancel")
	}
}
