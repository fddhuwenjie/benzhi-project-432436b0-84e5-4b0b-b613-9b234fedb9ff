package domain

import (
	"fmt"
	"strings"
	"time"
)

type ValidationIssue struct {
	Field   string
	Message string
}

func ValidateCase(c *TreatmentCase) []ValidationIssue {
	issues := []ValidationIssue{}
	if c == nil {
		return []ValidationIssue{{"case", "missing"}}
	}
	if strings.TrimSpace(c.CaseID) == "" {
		issues = append(issues, ValidationIssue{"case_id", "required"})
	}
	if strings.TrimSpace(c.SiteName) == "" {
		issues = append(issues, ValidationIssue{"site_name", "required"})
	}
	if strings.TrimSpace(c.MuralSection) == "" {
		issues = append(issues, ValidationIssue{"mural_section", "required"})
	}
	if strings.TrimSpace(c.OwnerID) == "" {
		issues = append(issues, ValidationIssue{"owner_id", "required"})
	}
	if !ValidStatus(c.Status) {
		issues = append(issues, ValidationIssue{"status", "unknown"})
	}
	if c.Revision < 1 {
		issues = append(issues, ValidationIssue{"revision", "positive"})
	}
	return issues
}
func ValidateAssessment(a ContaminationAssessment) []ValidationIssue {
	r := []ValidationIssue{}
	if a.AssessmentID == "" {
		r = append(r, ValidationIssue{"assessment_id", "required"})
	}
	if len(a.SamplePoints) == 0 {
		r = append(r, ValidationIssue{"sample_points", "at least one"})
	}
	if a.OrganismFindings == "" {
		r = append(r, ValidationIssue{"organism_findings", "required"})
	}
	if a.ActivityLevel == "" {
		r = append(r, ValidationIssue{"activity_level", "required"})
	}
	if a.SpreadBoundary == "" {
		r = append(r, ValidationIssue{"spread_boundary", "required"})
	}
	if a.Method == "" {
		r = append(r, ValidationIssue{"method", "required"})
	}
	if a.AssessorID == "" {
		r = append(r, ValidationIssue{"assessor_id", "required"})
	}
	return r
}
func ValidatePlan(p TreatmentPlan) []ValidationIssue {
	r := []ValidationIssue{}
	vals := map[string]string{"material_name": p.MaterialName, "compatibility_basis": p.CompatibilityBasis, "application_parameters": p.ApplicationParameters, "protection_measures": p.ProtectionMeasures, "rollback_conditions": p.RollbackConditions, "author_id": p.AuthorID}
	for k, v := range vals {
		if strings.TrimSpace(v) == "" {
			r = append(r, ValidationIssue{k, "required"})
		}
	}
	if p.Version < 1 {
		r = append(r, ValidationIssue{"version", "positive"})
	}
	return r
}
func ValidatePilot(p PilotTrial) []ValidationIssue {
	r := []ValidationIssue{}
	if p.ObservationDays <= 0 {
		r = append(r, ValidationIssue{"observation_days", "positive"})
	}
	if p.ColorThreshold <= 0 {
		r = append(r, ValidationIssue{"color_threshold", "positive"})
	}
	if p.BeforeActivity < 0 || p.AfterActivity < 0 {
		r = append(r, ValidationIssue{"activity", "nonnegative"})
	}
	return r
}
func ValidateOutcome(o OutcomeVerification) []ValidationIssue {
	r := []ValidationIssue{}
	if o.VerifiedBy == "" {
		r = append(r, ValidationIssue{"verified_by", "required"})
	}
	if o.ObservationDays <= 0 {
		r = append(r, ValidationIssue{"observation_days", "positive"})
	}
	if o.ActivityThreshold < 0 || o.ColorThreshold < 0 || o.StabilityThreshold < 0 {
		r = append(r, ValidationIssue{"thresholds", "nonnegative"})
	}
	return r
}
func (c *TreatmentCase) Summary() string {
	return fmt.Sprintf("%s/%s %s rev=%d", c.SiteName, c.MuralSection, c.Status, c.Revision)
}
func (c *TreatmentCase) IsTerminal() bool                { return c.Status == StatusArchived }
func (c *TreatmentCase) Age(now time.Time) time.Duration { return now.Sub(c.CreatedAt) }
