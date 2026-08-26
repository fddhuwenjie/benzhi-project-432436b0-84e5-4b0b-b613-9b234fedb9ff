package application

import (
	"mural-biocare/internal/domain"
	"mural-biocare/internal/persistence"
	"testing"
	"time"
)

func TestExtendedWorkflowArchivesIndexedEvidence(t *testing.T) {
	store, err := persistence.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	app := New(store)
	c, err := app.Create("现场", "A区", "菌斑", "owner", 20, 50, "create-1")
	if err != nil {
		t.Fatal(err)
	}
	a := domain.ContaminationAssessment{SamplePoints: []domain.SamplePoint{{ID: "s1", Location: "A区", Result: "positive", CollectedAt: time.Now()}}, OrganismFindings: "真菌", ActivityLevel: "HIGH", SpreadBoundary: "A区", Method: "culture", AssessorID: "assessor"}
	c, err = app.Assessment(c.CaseID, a, Command{Actor: "assessor", ExpectedRevision: c.Revision, Payload: a})
	if err != nil {
		t.Fatal(err)
	}
	p := domain.TreatmentPlan{MaterialName: "材料", CompatibilityBasis: "兼容", ApplicationParameters: "1%", ProtectionMeasures: "PPE", RollbackConditions: "冲洗", RequiredObservationDays: 3}
	c, err = app.Plan(c.CaseID, p, Command{Actor: "author", ExpectedRevision: c.Revision, Payload: p})
	if err != nil {
		t.Fatal(err)
	}
	items := []domain.PlanReviewItem{{Item: "compatibility_basis", Result: "PASS"}, {Item: "application_parameters", Result: "PASS"}, {Item: "protection_measures", Result: "PASS"}, {Item: "rollback_conditions", Result: "PASS"}}
	c, err = app.ReviewChecklist(c.CaseID, c.Plan.Version, "APPROVE", "", items, Command{Actor: "reviewer", ExpectedRevision: c.Revision, Payload: items})
	if err != nil {
		t.Fatal(err)
	}
	pTrial := domain.PilotTrial{ObservationDays: 3, ColorThreshold: 2, Observations: []domain.PilotObservation{{ObservationDay: 1, Activity: 10, ColorDelta: 1}, {ObservationDay: 2, Activity: 7, ColorDelta: 1}, {ObservationDay: 3, Activity: 4, ColorDelta: 1}}}
	c, err = app.Pilot(c.CaseID, pTrial, Command{Actor: "author", ExpectedRevision: c.Revision, Payload: pTrial})
	if err != nil {
		t.Fatal(err)
	}
	locked := []domain.PlannedCheckpoint{{CheckpointID: "cp-1", Sequence: 1, Phase: "TREATMENT", ExpectedCondition: "稳定"}}
	c, err = app.StartWithPlan(c.CaseID, locked, Command{Actor: "leader", ExpectedRevision: c.Revision, Payload: locked})
	if err != nil {
		t.Fatal(err)
	}
	cp := domain.ExecutionCheckpoint{CheckpointID: "cp-1", Sequence: 1, Phase: "TREATMENT", ExpectedCondition: "稳定", ObservedValue: "稳定", Result: "PASS", RecordedBy: "leader"}
	c, err = app.Checkpoint(c.CaseID, cp, Command{Actor: "leader", ExpectedRevision: c.Revision, Payload: cp})
	if err != nil {
		t.Fatal(err)
	}
	c, err = app.Complete(c.CaseID, Command{Actor: "leader", ExpectedRevision: c.Revision})
	if err != nil {
		t.Fatal(err)
	}
	o := domain.OutcomeVerification{PostActivity: 1, ColorDelta: 1, SurfaceStability: 0.9, ActivityThreshold: 2, ColorThreshold: 2, StabilityThreshold: 0.8, ObservationDays: 3, MeasuredAt: time.Now()}
	c, err = app.Outcome(c.CaseID, o, Command{Actor: "verifier", ExpectedRevision: c.Revision, Payload: o})
	if err != nil {
		t.Fatal(err)
	}
	ready, err := app.ArchiveReadiness(c.CaseID)
	if err != nil || !ready.Ready {
		t.Fatalf("预检未通过: %+v %v", ready, err)
	}
	c, err = app.Archive(c.CaseID, Command{Actor: "owner", ExpectedRevision: c.Revision})
	if err != nil {
		t.Fatal(err)
	}
	if err = app.Verify(c.CaseID); err != nil {
		t.Fatal(err)
	}
	if c.Evidence == nil || len(c.Evidence.EvidenceIndex) != 6 {
		t.Fatalf("证据索引不完整: %+v", c.Evidence)
	}
}
