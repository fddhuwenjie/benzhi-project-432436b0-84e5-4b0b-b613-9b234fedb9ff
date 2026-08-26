package persistence

import (
	"mural-biocare/internal/domain"
	"time"
)

func Clone(c *domain.TreatmentCase) *domain.TreatmentCase {
	if c == nil {
		return nil
	}
	cp := *c
	cp.BaselineUnlockedAt = cloneTime(c.BaselineUnlockedAt)
	cp.ArchivedAt = cloneTime(c.ArchivedAt)
	cp.BaselineMeasurements = append([]domain.BaselineMeasurement(nil), c.BaselineMeasurements...)
	cp.Plan = clonePlan(c.Plan)
	cp.PlanHistory = make([]domain.TreatmentPlan, len(c.PlanHistory))
	for i := range c.PlanHistory {
		cp.PlanHistory[i] = *clonePlan(&c.PlanHistory[i])
	}
	cp.Pilot = clonePilot(c.Pilot)
	cp.PilotHistory = make([]domain.PilotTrial, len(c.PilotHistory))
	for i := range c.PilotHistory {
		cp.PilotHistory[i] = *clonePilot(&c.PilotHistory[i])
	}
	cp.Checkpoints = cloneCheckpoints(c.Checkpoints)
	cp.CheckpointSummary = append([]string(nil), c.CheckpointSummary...)
	cp.CheckpointCompletionSummary = append([]domain.CheckpointCompletion(nil), c.CheckpointCompletionSummary...)
	cp.Outcome = cloneOutcome(c.Outcome)
	cp.OutcomeHistory = make([]domain.OutcomeVerification, len(c.OutcomeHistory))
	for i := range c.OutcomeHistory {
		cp.OutcomeHistory[i] = *cloneOutcome(&c.OutcomeHistory[i])
	}
	cp.Deviations = cloneDeviations(c.Deviations)
	cp.ProfileCorrections = append([]domain.ProfileCorrection(nil), c.ProfileCorrections...)
	for i := range cp.ProfileCorrections {
		cp.ProfileCorrections[i].Changes = append([]domain.FieldChange(nil), c.ProfileCorrections[i].Changes...)
	}
	cp.AssessmentDiffs = cloneAssessmentDiffs(c.AssessmentDiffs)
	cp.PlannedCheckpoints = append([]domain.PlannedCheckpoint(nil), c.PlannedCheckpoints...)
	cp.RectificationItems = append([]domain.RectificationItem(nil), c.RectificationItems...)
	cp.Evidence = cloneEvidence(c.Evidence)
	return &cp
}

func clonePlan(plan *domain.TreatmentPlan) *domain.TreatmentPlan {
	if plan == nil {
		return nil
	}
	cp := *plan
	cp.ChangedFields = append([]string(nil), plan.ChangedFields...)
	cp.ReviewItems = append([]domain.PlanReviewItem(nil), plan.ReviewItems...)
	return &cp
}

func clonePilot(pilot *domain.PilotTrial) *domain.PilotTrial {
	if pilot == nil {
		return nil
	}
	cp := *pilot
	cp.FailureReasons = append([]string(nil), pilot.FailureReasons...)
	cp.Observations = append([]domain.PilotObservation(nil), pilot.Observations...)
	return &cp
}

func cloneCheckpoints(checkpoints []domain.ExecutionCheckpoint) []domain.ExecutionCheckpoint {
	cp := append([]domain.ExecutionCheckpoint(nil), checkpoints...)
	for i := range cp {
		cp[i].EvidenceRefs = append([]string(nil), checkpoints[i].EvidenceRefs...)
	}
	return cp
}

func cloneOutcome(outcome *domain.OutcomeVerification) *domain.OutcomeVerification {
	if outcome == nil {
		return nil
	}
	cp := *outcome
	cp.FailureItems = append([]string(nil), outcome.FailureItems...)
	cp.ThresholdSnapshot = make(map[string]float64, len(outcome.ThresholdSnapshot))
	for key, value := range outcome.ThresholdSnapshot {
		cp.ThresholdSnapshot[key] = value
	}
	return &cp
}

func cloneDeviations(deviations []domain.Deviation) []domain.Deviation {
	cp := append([]domain.Deviation(nil), deviations...)
	for i := range cp {
		cp[i].CreatedAt = cloneTime(deviations[i].CreatedAt)
		cp[i].ResolvedAt = cloneTime(deviations[i].ResolvedAt)
		cp[i].EvidenceRefs = append([]string(nil), deviations[i].EvidenceRefs...)
	}
	return cp
}

func cloneAssessmentDiffs(diffs []domain.AssessmentDiff) []domain.AssessmentDiff {
	cp := append([]domain.AssessmentDiff(nil), diffs...)
	for i := range cp {
		cp[i].Added = append([]domain.SamplePoint(nil), diffs[i].Added...)
		cp[i].Removed = append([]domain.SamplePoint(nil), diffs[i].Removed...)
		cp[i].ChangedFields = append([]domain.FieldChange(nil), diffs[i].ChangedFields...)
		cp[i].Changed = append([]domain.SamplePointChange(nil), diffs[i].Changed...)
		for j := range cp[i].Changed {
			if diffs[i].Changed[j].Before != nil {
				before := *diffs[i].Changed[j].Before
				cp[i].Changed[j].Before = &before
			}
			if diffs[i].Changed[j].After != nil {
				after := *diffs[i].Changed[j].After
				cp[i].Changed[j].After = &after
			}
		}
	}
	return cp
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cp := *value
	return &cp
}

func cloneEvidence(evidence *domain.EvidenceManifest) *domain.EvidenceManifest {
	if evidence == nil {
		return nil
	}
	cp := *evidence
	cp.EvidenceItems = append([]string(nil), evidence.EvidenceItems...)
	cp.EvidenceIndex = make(map[string][]string, len(evidence.EvidenceIndex))
	for key, items := range evidence.EvidenceIndex {
		cp.EvidenceIndex[key] = append([]string(nil), items...)
	}
	return &cp
}
