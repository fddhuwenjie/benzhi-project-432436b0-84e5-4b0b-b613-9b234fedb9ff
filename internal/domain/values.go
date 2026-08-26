package domain

import "strings"

func NormalizeActivity(v string) string { return strings.ToUpper(strings.TrimSpace(v)) }
func ValidStatus(s Status) bool {
	switch s {
	case StatusDraft, StatusAssessed, StatusPlanApproved, StatusPilotPassed, StatusInProgress, StatusPaused, StatusTreatmentCompleted, StatusOutcomeVerified, StatusArchived, StatusPlanRevision, StatusPilotRevision:
		return true
	}
	return false
}
