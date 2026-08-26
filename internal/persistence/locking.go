package persistence

import "sync"

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
