package persistence

import (
	"context"
	"sync"
)

type CaseLocker struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

func NewCaseLocker() *CaseLocker { return &CaseLocker{locks: map[string]*sync.Mutex{}} }
func (l *CaseLocker) Lock(id string) func() {
	l.mu.Lock()
	m := l.locks[id]
	if m == nil {
		m = &sync.Mutex{}
		l.locks[id] = m
	}
	l.mu.Unlock()
	m.Lock()
	return m.Unlock
}

// LockContext acquires the per-case lock while respecting ctx cancellation.
// When another write command already holds the lock for the same case_id, the
// wait is interrupted as soon as ctx is canceled rather than blocking until the
// lock is released. On cancellation no lock is acquired and the returned done
// function is nil, so the caller must only invoke done when err is nil. A nil
// ctx behaves like context.Background(): the call blocks until the lock is free.
func (l *CaseLocker) LockContext(id string, ctx context.Context) (done func(), err error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	l.mu.Lock()
	m := l.locks[id]
	if m == nil {
		m = &sync.Mutex{}
		l.locks[id] = m
	}
	l.mu.Unlock()

	acquired := make(chan struct{})
	released := make(chan struct{})
	go func() {
		m.Lock()
		close(acquired)
		<-released
		m.Unlock()
	}()
	if ctx == nil {
		<-acquired
		return func() { close(released) }, nil
	}
	select {
	case <-acquired:
		return func() { close(released) }, nil
	case <-ctx.Done():
		close(released)
		return nil, ctx.Err()
	}
}
