package persistence

func (s *Store) Healthy() bool { return s.Validate() == nil }
