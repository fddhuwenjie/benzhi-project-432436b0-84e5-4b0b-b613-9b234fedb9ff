package audit_event_orphan_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"mural-biocare/internal/audit"
	"mural-biocare/internal/domain"
	"mural-biocare/internal/persistence"
)

func TestFailedSnapshotCommitDoesNotPublishAuditEvent(t *testing.T) {
	dir := t.TempDir()
	store, err := persistence.New(dir)
	if err != nil {
		t.Fatal(err)
	}

	const caseID = "case-orphan-commit"
	c, err := domain.NewCase(caseID, "莫高窟", "A-01", "菌斑", "owner-1", 20, 50, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	created := audit.Append(nil, caseID, c.Revision, "CASE_CREATED", c.OwnerID, c, time.Now())
	if err := store.Save(c, created); err != nil {
		t.Fatal(err)
	}

	revised, ok := store.Get(caseID)
	if !ok {
		t.Fatal("saved case not found")
	}
	nextSymptom := "菌斑范围扩大"
	if err := revised.CorrectProfile(domain.ProfileCorrectionInput{
		SymptomDescription: &nextSymptom,
		Reason:             "现场复核",
	}, c.OwnerID, c.Revision); err != nil {
		t.Fatal(err)
	}
	corrected := audit.Append(store.Events(caseID), caseID, revised.Revision, "PROFILE_CORRECTED", c.OwnerID, revised.ProfileCorrections[0], time.Now())

	// 将已提交快照替换为同名目录，使下一次快照发布必然失败，但不妨碍审计文件替换。
	snapshotPath := filepath.Join(dir, caseID+".json")
	if err := os.Remove(snapshotPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(snapshotPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(revised, corrected); err == nil {
		t.Fatal("expected snapshot commit failure")
	}

	reloaded, err := persistence.New(dir)
	if err != nil {
		t.Fatal(err)
	}
	events := reloaded.Events(caseID)
	if len(events) != 1 || events[0].Digest != created.Digest {
		t.Fatalf("failed snapshot commit replaced prior audit chain with %d event(s)", len(events))
	}
}
