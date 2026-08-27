package restart_quarantine_scope_test

import (
	"mural-biocare/internal/application"
	"mural-biocare/internal/persistence"
	"os"
	"path/filepath"
	"testing"
)

func TestRestartQuarantinesOnlyCorruptedCase(t *testing.T) {
	dir := t.TempDir()
	store, err := persistence.New(dir)
	if err != nil {
		t.Fatal(err)
	}
	app := application.New(store)
	healthy, err := app.Create("一号窟", "东壁", "局部菌斑", "owner-1", 20, 55, "healthy-create")
	if err != nil {
		t.Fatal(err)
	}

	corruptEvents := filepath.Join(dir, "corrupted-case.events.jsonl")
	if err := os.WriteFile(corruptEvents, []byte("{not-json}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	restarted, err := persistence.New(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := restarted.Get(healthy.CaseID); !ok {
		t.Fatalf("损坏个案的隔离扩大到健康个案 %s", healthy.CaseID)
	}
	if _, ok := restarted.Get("corrupted-case"); ok {
		t.Fatal("损坏个案不应进入仓储")
	}
}
