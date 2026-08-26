package persistence

import "mural-biocare/internal/domain"

func Clone(c *domain.TreatmentCase) *domain.TreatmentCase {
	if c == nil {
		return nil
	}
	cp := *c
	cp.Checkpoints = append([]domain.ExecutionCheckpoint(nil), c.Checkpoints...)
	cp.Deviations = append([]domain.Deviation(nil), c.Deviations...)
	return &cp
}
