// Package worker provides coalesced wakeups for durable database-backed work.
package worker

import (
	"context"
	"sync"
	"time"
)

const ReconcileInterval = time.Minute

// Signal is a hint to inspect durable state, not a queue of individual jobs.
// Its zero value is ready to use. Signals received while busy are coalesced.
type Signal struct {
	once sync.Once
	wake chan struct{}
}

func (s *Signal) C() <-chan struct{} {
	s.once.Do(func() { s.wake = make(chan struct{}, 1) })
	return s.wake
}

func (s *Signal) Notify() {
	s.C()
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

// Run checks on startup, drains successful batches, then waits for a signal or
// the reconciliation interval. Errors leave work durable for a later attempt.
func Run(
	ctx context.Context,
	wake <-chan struct{},
	interval time.Duration,
	process func(context.Context) (bool, error),
	report func(error),
) {
	timer := time.NewTimer(interval)
	defer timer.Stop()
	for ctx.Err() == nil {
		more, err := process(ctx)
		if err != nil && ctx.Err() == nil {
			report(err)
		}
		if more && err == nil {
			continue
		}
		timer.Reset(interval)
		select {
		case <-ctx.Done():
			return
		case <-wake:
		case <-timer.C:
		}
	}
}
