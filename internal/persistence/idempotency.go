package persistence

func (s *Store) ClearIdempotency() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.idem = map[string]Idempotent{}
}
