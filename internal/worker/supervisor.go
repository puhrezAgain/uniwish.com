/*
uniwish.com/interal/worker/supervisor

contains the logic for running the worker loop
*/
package worker

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"time"
)

type JobWorker interface {
	RunOnce(context.Context) error
}
type WorkerSupervisor struct {
	Worker           JobWorker
	PollInterval     time.Duration
	FailureTolerance int
	Sleep            func(time.Duration)
	OnFatal          func()
	Logger           *slog.Logger
}

func (ws *WorkerSupervisor) Run(ctx context.Context) {
	var failures atomic.Int32

	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := ws.Worker.RunOnce(ctx)
			var je JobError
			switch {
			case err == nil:
				failures.Store(0)
			case errors.Is(err, ErrNoWork):
				// No work is not a failure nor a success
				// so lets not reset the failure counter, but also not increment it
			case errors.As(err, &je):
				// handling dead letter is proper health
				// resetting the failure counter because processing problems could be input issues
				ws.Logger.Error("job error", "error", err)
				failures.Store(0)
			default:
				ws.Logger.Error("worker error", "error", err)

				if failures.Add(1) >= int32(ws.FailureTolerance) {
					ws.Logger.Error("worker tolerance exceeded", "failures", failures.Load())
					ws.OnFatal()
					return
				}
			}
			ws.Sleep(ws.PollInterval)
		}
	}

}
