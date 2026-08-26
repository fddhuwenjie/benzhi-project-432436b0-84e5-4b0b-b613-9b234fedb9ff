package domain

import (
	"errors"
	"testing"
	"time"
)

func TestProfileCorrectionAndAssessmentVersions(t *testing.T) {
	now := time.Now().Add(-time.Hour)
	c, err := NewCase("case-1", "现场", "一区", "霉斑", "owner-1", 20, 50, now)
	if err != nil {
		t.Fatal(err)
	}
	section := "二区"
	if err = c.CorrectProfile(ProfileCorrectionInput{MuralSection: &section, Reason: "纠正录入区段"}, "owner-1", c.Revision); err != nil {
		t.Fatal(err)
	}
	if c.Revision != 2 || c.MuralSection != section || len(c.ProfileCorrections) != 1 {
		t.Fatalf("更正未原子保存: %+v", c)
	}
	rev := c.Revision
	if err = c.CorrectProfile(ProfileCorrectionInput{MuralSection: &section, Reason: "重复"}, "owner-1", rev); err == nil || c.Revision != rev {
		t.Fatal("无变化更正不应写入")
	}
	a1 := ContaminationAssessment{SamplePoints: []SamplePoint{{ID: "s1", Location: "二区", Result: "positive", CollectedAt: now.Add(10 * time.Minute)}}, OrganismFindings: "真菌", ActivityLevel: "high", SpreadBoundary: "二区", Method: "culture", AssessorID: "assessor"}
	if err = c.SubmitAssessment(a1, c.Revision); err != nil {
		t.Fatal(err)
	}
	a2 := a1
	a2.SamplePoints = []SamplePoint{{ID: "s1", Location: "二区", Result: "positive", CollectedAt: now.Add(20 * time.Minute)}, {ID: "s2", Location: "二区", Result: "positive", CollectedAt: now.Add(30 * time.Minute)}}
	a2.SpreadBoundary = "二区东侧"
	if err = c.SubmitAssessment(a2, c.Revision); err != nil {
		t.Fatal(err)
	}
	d := BuildAssessmentDiff(c.AssessmentHistory[0], *c.Assessment)
	if d.FromVersion != 1 || d.ToVersion != 2 || len(d.Added) != 1 || len(d.Changed) != 1 {
		t.Fatalf("评估差异错误: %+v", d)
	}
}

func TestReviewChecklistAndPilotObservationGate(t *testing.T) {
	c, _ := NewCase("case-2", "现场", "一区", "霉斑", "owner", 20, 50, time.Now().Add(-time.Hour))
	c.Status = StatusAssessed
	c.Plan = &TreatmentPlan{Version: 1, AuthorID: "author", RequiredObservationDays: 3}
	items := []PlanReviewItem{{Item: "compatibility_basis", Result: "PASS"}, {Item: "application_parameters", Result: "PASS"}, {Item: "protection_measures", Result: "FAIL", Comment: "补充防护面罩"}, {Item: "rollback_conditions", Result: "PASS"}}
	rev := c.Revision
	if err := c.ReviewPlanChecklist("reviewer", "APPROVE", "", items, 1, rev); err == nil || c.Revision != rev {
		t.Fatal("FAIL 清单不应批准")
	}
	items[2].Result, items[2].Comment = "PASS", ""
	if err := c.ReviewPlanChecklist("reviewer", "APPROVE", "", items, 1, rev); err != nil {
		t.Fatal(err)
	}
	p := PilotTrial{ObservationDays: 3, ColorThreshold: 2, Observations: []PilotObservation{{1, 10, 1}, {2, 7, 3}, {3, 4, 1}}}
	if err := c.SubmitPilot(p, c.Revision); err != nil {
		t.Fatal(err)
	}
	if c.Pilot.Passed || c.Status != StatusPilotRevision || len(c.Pilot.FailureReasons) == 0 {
		t.Fatalf("中间色差越限未拦截: %+v", c.Pilot)
	}
}

func TestLockedPlanRecurrenceAndRectification(t *testing.T) {
	c, _ := NewCase("case-3", "现场", "一区", "霉斑", "owner", 20, 50, time.Now().Add(-time.Hour))
	c.Status = StatusPilotPassed
	plan := []PlannedCheckpoint{{CheckpointID: "cp-1", Sequence: 1, Phase: "TREATMENT", ExpectedCondition: "无污染", Unit: ""}, {CheckpointID: "cp-2", Sequence: 2, Phase: "TREATMENT", ExpectedCondition: "无污染", Unit: ""}, {CheckpointID: "cp-3", Sequence: 3, Phase: "TREATMENT", ExpectedCondition: "无污染", Unit: ""}}
	if err := c.StartExecutionWithPlan("leader", plan, c.Revision); err != nil {
		t.Fatal(err)
	}
	rev := c.Revision
	bad := ExecutionCheckpoint{CheckpointID: "cp-x", Sequence: 1, Phase: "TREATMENT", ExpectedCondition: "无污染", ObservedValue: "ok", Result: "PASS", RecordedBy: "worker"}
	if err := c.RecordCheckpoint(bad, rev); !errors.Is(err, ErrConflict) || c.Revision != rev {
		t.Fatal("篡改锁定计划未返回冲突")
	}
	first := ExecutionCheckpoint{CheckpointID: "cp-1", Sequence: 1, Phase: "TREATMENT", ExpectedCondition: "无污染", ObservedValue: "minor", Result: "DEVIATION", DeviationID: "d1", DeviationType: "PROCESS", RecordedBy: "worker"}
	if err := c.RecordCheckpoint(first, rev); err != nil {
		t.Fatal(err)
	}
	if err := c.ResolveDeviationWithEvidence("d1", "纠正", []string{"e1"}, "", "leader", c.Revision); err != nil {
		t.Fatal(err)
	}
	second := ExecutionCheckpoint{CheckpointID: "cp-2", Sequence: 2, Phase: "TREATMENT", ExpectedCondition: "无污染", ObservedValue: "minor", Result: "DEVIATION", DeviationID: "d2", DeviationType: "PROCESS", RelatedDeviationID: "d1", RecordedBy: "worker"}
	if err := c.RecordCheckpoint(second, c.Revision); err != nil {
		t.Fatal(err)
	}
	if c.Deviations[1].Severity != "MEDIUM" || c.Deviations[1].RootDeviationID != "d1" {
		t.Fatalf("首次复发未升级: %+v", c.Deviations[1])
	}
	if err := c.ResolveDeviationWithEvidence("d2", "纠正", []string{"e2"}, "", "leader", c.Revision); err != nil {
		t.Fatal(err)
	}
	third := ExecutionCheckpoint{CheckpointID: "cp-3", Sequence: 3, Phase: "TREATMENT", ExpectedCondition: "无污染", ObservedValue: "minor", Result: "DEVIATION", DeviationID: "d3", DeviationType: "PROCESS", RelatedDeviationID: "d2", RecordedBy: "worker"}
	if err := c.RecordCheckpoint(third, c.Revision); err != nil {
		t.Fatal(err)
	}
	if c.Deviations[2].Severity != "HIGH" {
		t.Fatalf("再次复发未升级 HIGH: %+v", c.Deviations[2])
	}
	rev = c.Revision
	if err := c.ResolveDeviationWithEvidence("d3", "纠正", []string{"e3"}, "reviewer", "leader", rev); err == nil || c.Revision != rev {
		t.Fatal("HIGH 偏差不应接受单条证据")
	}
}
