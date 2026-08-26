package audit

import "mural-biocare/internal/domain"

func EvidenceCount(c *domain.TreatmentCase) int {
	if c.Evidence == nil {
		return 0
	}
	return len(c.Evidence.EvidenceItems)
}
