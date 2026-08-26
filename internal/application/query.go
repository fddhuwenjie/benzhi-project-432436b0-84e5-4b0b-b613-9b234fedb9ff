package application

import "mural-biocare/internal/domain"

type CaseView struct {
	Case          *domain.TreatmentCase
	TimelineLen   int
	ManifestValid bool
}

func (s *Service) View(id string) (CaseView, error) {
	c, ok := s.Get(id)
	if !ok {
		return CaseView{}, ErrNotFound
	}
	return CaseView{Case: c, TimelineLen: len(s.Timeline(id)), ManifestValid: s.Verify(id) == nil}, nil
}

var ErrNotFound = errorString("not found")

type errorString string

func (e errorString) Error() string { return string(e) }
