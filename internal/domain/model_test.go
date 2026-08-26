package domain

import (
	"testing"
	"time"
)

func TestLifecycle(t *testing.T) {
	c, e := NewCase("c", "site", "sec", "sym", "owner", 20, 50, time.Now())
	if e != nil {
		t.Fatal(e)
	}
	if e = c.SubmitAssessment(ContaminationAssessment{SamplePoints: []SamplePoint{{ID: "1", Location: "sec", Result: "positive", CollectedAt: time.Now()}}, OrganismFindings: "x", ActivityLevel: "high", SpreadBoundary: "s", Method: "culture", AssessorID: "d"}, 1); e != nil {
		t.Fatal(e)
	}
	if e = c.SubmitPlan(TreatmentPlan{MaterialName: "m", CompatibilityBasis: "b", ApplicationParameters: "p", ProtectionMeasures: "ppe", RollbackConditions: "r", AuthorID: "a", RequiredObservationDays: 7}, 2); e != nil {
		t.Fatal(e)
	}
	if e = c.ReviewPlan("r", "APPROVE", "", 3); e != nil {
		t.Fatal(e)
	}
	if c.Status != StatusPlanApproved {
		t.Fatal(c.Status)
	}
}
func TestRevisionConflict(t *testing.T) {
	c, _ := NewCase("c", "s", "m", "x", "o", 20, 50, time.Now())
	if e := c.StartExecution("x", 99); e == nil {
		t.Fatal("expected conflict")
	}
}

func TestBaselineRetestGate(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	c, err := NewCase("c", "site", "sec", "sym", "owner", 40, 50, now)
	if err != nil {
		t.Fatal(err)
	}
	a := ContaminationAssessment{SamplePoints: []SamplePoint{{ID: "s1", Location: "sec", Result: "positive", CollectedAt: now}}, OrganismFindings: "fungus", ActivityLevel: "HIGH", SpreadBoundary: "sec", Method: "culture", AssessorID: "detector"}
	if err = c.SubmitAssessment(a, c.Revision); err == nil || c.Revision != 1 {
		t.Fatalf("异常基线不应通过评估: %v", err)
	}
	if err = c.UpdateBaseline(20, 50, now.Add(24*time.Hour), "operator", c.Revision); err != nil {
		t.Fatal(err)
	}
	if c.BaselineStatus != BaselineRetestRequired {
		t.Fatal("单次正常读数不应解锁")
	}
	if err = c.UpdateBaseline(21, 51, now.Add(48*time.Hour), "operator", c.Revision); err != nil {
		t.Fatal(err)
	}
	if c.BaselineStatus != BaselineNormal || c.BaselineUnlockedAt == nil || c.BaselineRetestCount != 2 {
		t.Fatalf("复测未解锁: %+v", c)
	}
	if err = c.SubmitAssessment(a, c.Revision); err != nil {
		t.Fatal(err)
	}
}

func TestCheckpointBatchIsAtomicAndHighDeviationNeedsEvidence(t *testing.T) {
	c, _ := NewCase("c", "site", "sec", "sym", "owner", 20, 50, time.Now())
	c.Status = StatusInProgress
	bad := []ExecutionCheckpoint{
		{Sequence: 1, Phase: "TREATMENT", ExpectedCondition: "ok", ObservedValue: "ok", Result: "PASS", RecordedBy: "worker"},
		{Sequence: 3, Phase: "TREATMENT", ExpectedCondition: "ok", ObservedValue: "ok", Result: "PASS", RecordedBy: "worker"},
	}
	if err := c.RecordCheckpoints(bad, c.Revision); err == nil || len(c.Checkpoints) != 0 || c.Revision != 1 {
		t.Fatalf("无效批次发生部分写入: %v", err)
	}
	batch := []ExecutionCheckpoint{
		{Sequence: 1, Phase: "TREATMENT", ExpectedCondition: "ok", ObservedValue: "ok", Result: "PASS", RecordedBy: "worker"},
		{Sequence: 2, Phase: "APPLICATION", ExpectedCondition: "safe", ObservedValue: "spill", Result: "DEVIATION", DeviationID: "d1", DeviationType: "SAFETY", RecordedBy: "worker"},
	}
	if err := c.RecordCheckpoints(batch, c.Revision); err != nil {
		t.Fatal(err)
	}
	if c.Status != StatusPaused || c.Deviations[0].Severity != "HIGH" {
		t.Fatalf("偏差未暂停或分级错误: %+v", c.Deviations)
	}
	rev := c.Revision
	if err := c.ResolveDeviationWithEvidence("d1", "clean", []string{"ev-1"}, "reviewer", "leader", rev); err == nil || c.Revision != rev {
		t.Fatal("HIGH 偏差不应接受单条证据")
	}
	if err := c.ResolveDeviationWithEvidence("d1", "clean", []string{"ev-1", "ev-2"}, "reviewer", "leader", rev); err != nil {
		t.Fatal(err)
	}
	if c.Status != StatusInProgress || c.Checkpoints[1].Result != "PASS" {
		t.Fatal("解决全部偏差后未恢复")
	}
	rev = c.Revision
	if err := c.ResolveDeviationWithEvidence("d1", "again", []string{"ev-3", "ev-4"}, "reviewer-2", "leader", rev); err == nil || c.Revision != rev {
		t.Fatal("重复解决应返回错误且不改变修订号")
	}
}
