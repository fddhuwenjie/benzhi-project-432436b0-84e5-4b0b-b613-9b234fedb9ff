package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type ProfileCorrectionInput struct {
	SiteName           *string
	MuralSection       *string
	SymptomDescription *string
	OwnerID            *string
	Reason             string
}

func (c *TreatmentCase) CorrectProfile(in ProfileCorrectionInput, actor string, rev int) error {
	if err := c.check(rev); err != nil {
		return err
	}
	if c.Status != StatusDraft {
		return fmt.Errorf("%w: profile correction only allowed in DRAFT", ErrForbidden)
	}
	if strings.TrimSpace(actor) == "" || actor != c.OwnerID {
		return fmt.Errorf("%w: only current owner may correct profile", ErrInvalid)
	}
	if strings.TrimSpace(in.Reason) == "" {
		return fmt.Errorf("%w: correction_reason required", ErrInvalid)
	}
	site, section, symptom, owner := c.SiteName, c.MuralSection, c.SymptomDescription, c.OwnerID
	if in.SiteName != nil {
		site = strings.TrimSpace(*in.SiteName)
	}
	if in.MuralSection != nil {
		section = strings.TrimSpace(*in.MuralSection)
	}
	if in.SymptomDescription != nil {
		symptom = strings.TrimSpace(*in.SymptomDescription)
	}
	if in.OwnerID != nil {
		owner = strings.TrimSpace(*in.OwnerID)
	}
	if site == "" || section == "" || symptom == "" || owner == "" {
		return fmt.Errorf("%w: corrected profile contains empty required field", ErrInvalid)
	}
	if in.OwnerID != nil && owner == c.OwnerID {
		return fmt.Errorf("%w: new owner must differ from current owner", ErrInvalid)
	}
	changes := make([]FieldChange, 0, 4)
	add := func(field, old, next string) {
		if old != next {
			changes = append(changes, FieldChange{Field: field, OldValue: old, NewValue: next})
		}
	}
	add("site_name", c.SiteName, site)
	add("mural_section", c.MuralSection, section)
	add("symptom_description", c.SymptomDescription, symptom)
	add("owner_id", c.OwnerID, owner)
	if len(changes) == 0 {
		return fmt.Errorf("%w: profile correction has no actual change", ErrInvalid)
	}
	previousOwner := c.OwnerID
	c.SiteName, c.MuralSection, c.SymptomDescription, c.OwnerID = site, section, symptom, owner
	c.bump()
	c.ProfileCorrections = append(c.ProfileCorrections, ProfileCorrection{Revision: c.Revision, Reason: strings.TrimSpace(in.Reason), Actor: actor, PreviousOwner: previousOwner, NewOwner: owner, Changes: changes, CorrectedAt: time.Now()})
	return nil
}

func BuildAssessmentDiff(from, to ContaminationAssessment) AssessmentDiff {
	d := AssessmentDiff{FromVersion: from.Version, ToVersion: to.Version}
	a, b := map[string]SamplePoint{}, map[string]SamplePoint{}
	for _, p := range from.SamplePoints {
		a[p.ID] = p
	}
	for _, p := range to.SamplePoints {
		b[p.ID] = p
	}
	for id, old := range a {
		next, ok := b[id]
		if !ok {
			d.Removed = append(d.Removed, old)
			continue
		}
		if old.Location != next.Location || old.Result != next.Result || !old.CollectedAt.Equal(next.CollectedAt) {
			x, y := old, next
			d.Changed = append(d.Changed, SamplePointChange{ID: id, Before: &x, After: &y})
		}
	}
	for id, next := range b {
		if _, ok := a[id]; !ok {
			d.Added = append(d.Added, next)
		}
	}
	sort.Slice(d.Added, func(i, j int) bool { return d.Added[i].ID < d.Added[j].ID })
	sort.Slice(d.Removed, func(i, j int) bool { return d.Removed[i].ID < d.Removed[j].ID })
	sort.Slice(d.Changed, func(i, j int) bool { return d.Changed[i].ID < d.Changed[j].ID })
	field := func(name, x, y string) {
		if x != y {
			d.ChangedFields = append(d.ChangedFields, FieldChange{Field: name, OldValue: x, NewValue: y})
		}
	}
	field("method", from.Method, to.Method)
	field("activity_level", from.ActivityLevel, to.ActivityLevel)
	field("spread_boundary", from.SpreadBoundary, to.SpreadBoundary)
	field("organism_findings", from.OrganismFindings, to.OrganismFindings)
	return d
}

func (c *TreatmentCase) AssessmentVersion(version int) (ContaminationAssessment, bool) {
	if c.Assessment != nil && c.Assessment.Version == version {
		return *c.Assessment, true
	}
	for _, a := range c.AssessmentHistory {
		if a.Version == version {
			return a, true
		}
	}
	return ContaminationAssessment{}, false
}

var reviewItemOrder = []string{"MATERIAL_COMPATIBILITY", "APPLICATION_PARAMETERS", "PERSONNEL_PROTECTION", "ROLLBACK_CONDITIONS"}

func normalizeReviewItems(items []PlanReviewItem, decision string) ([]PlanReviewItem, error) {
	if len(items) != len(reviewItemOrder) {
		return nil, fmt.Errorf("%w: exactly four review_items required", ErrInvalid)
	}
	aliases := map[string]string{"MATERIAL_COMPATIBILITY": "MATERIAL_COMPATIBILITY", "APPLICATION_PARAMETERS": "APPLICATION_PARAMETERS", "PERSONNEL_PROTECTION": "PERSONNEL_PROTECTION", "ROLLBACK_CONDITIONS": "ROLLBACK_CONDITIONS", "material_compatibility": "MATERIAL_COMPATIBILITY", "compatibility_basis": "MATERIAL_COMPATIBILITY", "application_parameters": "APPLICATION_PARAMETERS", "personnel_protection": "PERSONNEL_PROTECTION", "protection_measures": "PERSONNEL_PROTECTION", "rollback_conditions": "ROLLBACK_CONDITIONS"}
	byName := map[string]PlanReviewItem{}
	hasFail := false
	for _, item := range items {
		name, ok := aliases[strings.TrimSpace(item.Item)]
		if !ok {
			return nil, fmt.Errorf("%w: unknown review item %s", ErrInvalid, item.Item)
		}
		if _, exists := byName[name]; exists {
			return nil, fmt.Errorf("%w: duplicate review item %s", ErrInvalid, name)
		}
		item.Item, item.Result, item.Comment = name, strings.ToUpper(strings.TrimSpace(item.Result)), strings.TrimSpace(item.Comment)
		if item.Result != "PASS" && item.Result != "FAIL" {
			return nil, fmt.Errorf("%w: review item result", ErrInvalid)
		}
		if item.Result == "FAIL" {
			hasFail = true
			if item.Comment == "" {
				return nil, fmt.Errorf("%w: failed review item requires actionable comment", ErrInvalid)
			}
		}
		byName[name] = item
	}
	decision = strings.ToUpper(strings.TrimSpace(decision))
	if (decision == "APPROVE" && hasFail) || (decision == "REJECT" && !hasFail) {
		return nil, fmt.Errorf("%w: decision conflicts with review_items", ErrInvalid)
	}
	ordered := make([]PlanReviewItem, 0, 4)
	for _, name := range reviewItemOrder {
		ordered = append(ordered, byName[name])
	}
	return ordered, nil
}

func validCheckpointPhase(v string) bool {
	switch v {
	case "PREPARATION", "APPLICATION", "OBSERVATION", "TREATMENT", "RECTIFICATION":
		return true
	}
	return false
}

func validUnit(v string) bool {
	if v == "" {
		return true
	}
	switch v {
	case "C", "°C", "%", "lux", "mm", "mg/L", "min":
		return true
	}
	return false
}

func (c *TreatmentCase) openRectification(id string) (int, bool) {
	for i := range c.RectificationItems {
		if c.RectificationItems[i].ID == id && c.RectificationItems[i].Status == "OPEN" {
			return i, true
		}
	}
	return -1, false
}
