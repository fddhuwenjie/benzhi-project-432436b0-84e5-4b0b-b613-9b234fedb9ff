package application

import (
	"fmt"
	"mural-biocare/internal/domain"
)

type Gate struct {
	Name   string
	Ready  bool
	Detail string
}

func Gates(c *domain.TreatmentCase) []Gate {
	g := []Gate{{"assessment", c.Assessment != nil, "采样评估"}, {"plan", c.Plan != nil && c.Plan.ReviewDecision == "APPROVE", "方案复核"}, {"pilot", c.Pilot != nil && c.Pilot.Passed, "小区试验"}, {"execution", len(c.Checkpoints) > 0 && c.Status != domain.StatusPaused, "现场检查点"}, {"outcome", c.Outcome != nil && c.Outcome.Passed, "成效复验"}}
	return g
}
func Pending(c *domain.TreatmentCase) []string {
	r := []string{}
	for _, g := range Gates(c) {
		if !g.Ready {
			r = append(r, g.Name)
		}
	}
	return r
}
func EnsureActor(actor string) error {
	if actor == "" {
		return ErrUnauthorized
	}
	return nil
}
func EnsureRevision(c *domain.TreatmentCase, rev int) error {
	if c == nil {
		return fmt.Errorf("case missing")
	}
	if c.Revision != rev {
		return domain.ErrConflict
	}
	return nil
}
func NextAction(c *domain.TreatmentCase) string {
	switch c.Status {
	case domain.StatusDraft:
		return "assessment"
	case domain.StatusAssessed, domain.StatusPlanRevision, domain.StatusPilotRevision:
		return "plan_or_review"
	case domain.StatusPlanApproved:
		return "pilot"
	case domain.StatusPilotPassed:
		return "start"
	case domain.StatusInProgress:
		return "checkpoint_or_complete"
	case domain.StatusPaused:
		return "resolve"
	case domain.StatusTreatmentCompleted:
		return "outcome"
	case domain.StatusOutcomeVerified:
		return "archive"
	default:
		return "read_only"
	}
}
func Describe(c *domain.TreatmentCase) map[string]any {
	return map[string]any{"case_id": c.CaseID, "status": c.Status, "revision": c.Revision, "next_action": NextAction(c), "pending_gates": Pending(c), "archived": c.IsTerminal()}
}
