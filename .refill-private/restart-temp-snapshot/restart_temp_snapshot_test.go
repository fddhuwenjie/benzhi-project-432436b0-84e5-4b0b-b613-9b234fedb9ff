package restart_temp_snapshot_test

import (
	"mural-biocare/internal/audit"
	"mural-biocare/internal/domain"
	"mural-biocare/internal/persistence"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRestartIgnoresUncommittedSnapshotTemporaryFile(t *testing.T) {
	dir := t.TempDir()
	store, err := persistence.New(dir)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}

	createdAt := time.Date(2026, time.August, 26, 10, 0, 0, 0, time.UTC)
	c, err := domain.NewCase("case-temp-recovery", "一号窟", "东壁中段", "颜料层存在菌斑", "owner-1", 22.5, 55, createdAt)
	if err != nil {
		t.Fatalf("create case: %v", err)
	}
	event := audit.Append(nil, c.CaseID, c.Revision, "CASE_CREATED", c.OwnerID, c, createdAt)
	if err := store.Save(c, event); err != nil {
		t.Fatalf("save committed case: %v", err)
	}

	temporaryPath := filepath.Join(dir, c.CaseID+".snapshot.tmp")
	if err := os.WriteFile(temporaryPath, []byte(`{"case_id":`), 0644); err != nil {
		t.Fatalf("write interrupted temporary snapshot: %v", err)
	}

	restarted, err := persistence.New(dir)
	if err != nil {
		t.Fatalf("restart store: %v", err)
	}
	got, ok := restarted.Get(c.CaseID)
	if !ok {
		t.Fatalf("committed case disappeared after restart")
	}
	if got.Revision != c.Revision || got.SiteName != c.SiteName {
		t.Fatalf("restart loaded uncommitted state: revision=%d site_name=%q", got.Revision, got.SiteName)
	}
	if events := restarted.Events(c.CaseID); len(events) != 1 || events[0].Digest != event.Digest {
		t.Fatalf("committed audit history changed after restart: %#v", events)
	}
}
