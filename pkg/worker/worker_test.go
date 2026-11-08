package worker

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestRunStartupWakeReconcileAndShutdown(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		var signal Signal
		var calls atomic.Int64
		go Run(ctx, signal.C(), time.Minute, func(context.Context) (bool, error) {
			calls.Add(1)
			return false, nil
		}, func(err error) { t.Error(err) })
		synctest.Wait()
		if calls.Load() != 1 {
			t.Fatal("durable backlog was not inspected on startup")
		}
		time.Sleep(59 * time.Second)
		if calls.Load() != 1 {
			t.Fatal("idle worker polled before the fallback interval")
		}
		signal.Notify()
		synctest.Wait()
		if calls.Load() != 2 {
			t.Fatal("notification did not wake the worker immediately")
		}
		time.Sleep(time.Minute)
		synctest.Wait()
		if calls.Load() != 3 {
			t.Fatal("missed notification was not covered by reconciliation")
		}
		cancel()
		synctest.Wait()
		signal.Notify()
		time.Sleep(time.Minute)
		if calls.Load() != 3 {
			t.Fatal("worker queried after shutdown")
		}
	})
}

func TestRunDrainsBatchesButWaitsAfterErrors(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		var calls, failures atomic.Int64
		go Run(ctx, nil, time.Minute, func(context.Context) (bool, error) {
			if calls.Add(1) <= 3 {
				return true, nil
			}
			return true, errors.New("database unavailable")
		}, func(error) { failures.Add(1) })
		synctest.Wait()
		if calls.Load() != 4 || failures.Load() != 1 {
			t.Fatal("batches should drain immediately and stop at the first failure")
		}
		time.Sleep(time.Minute)
		synctest.Wait()
		if calls.Load() != 5 || failures.Load() != 2 {
			t.Fatal("failed work spun or was never retried")
		}
	})
}

func TestSignalCoalescesConcurrentWakeups(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var signal Signal
		for range 100 {
			go signal.Notify()
		}
		synctest.Wait()
		select {
		case <-signal.C():
		default:
			t.Fatal("wake was lost")
		}
		select {
		case <-signal.C():
			t.Fatal("hints must coalesce instead of building an unbounded queue")
		default:
		}
	})
}
