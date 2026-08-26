package application

import "mural-biocare/internal/audit"

func EventCount(s *Service, id string) int { return len(s.Timeline(id)) }
func LastEvent(s *Service, id string) *audit.Event {
	e := s.Timeline(id)
	if len(e) == 0 {
		return nil
	}
	return &e[len(e)-1]
}
