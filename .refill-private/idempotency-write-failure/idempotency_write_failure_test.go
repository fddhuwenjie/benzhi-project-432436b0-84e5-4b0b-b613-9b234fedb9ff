package idempotencywritefailure_test

import (
	"mural-biocare/internal/application"
	"mural-biocare/internal/domain"
	"mural-biocare/internal/persistence"
	"os"
	"path/filepath"
	"testing"
)

func TestMutationFailsWhenIdempotencyRecordCannotPersist(t *testing.T) {
	dir := t.TempDir()
	store, err := persistence.New(dir)
	if err != nil {
		t.Fatal(err)
	}
	app := application.New(store)
	created, err := app.Create("原现场", "A区", "菌斑", "owner", 20, 50, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "idempotency.json"), 0755); err != nil {
		t.Fatal(err)
	}
	newSite := "新现场"
	input := domain.ProfileCorrectionInput{SiteName: &newSite, Reason: "纠正现场"}
	_, err = app.CorrectProfile(created.CaseID, input, application.Command{
		RequestID:        "profile-request",
		Actor:            "owner",
		ExpectedRevision: created.Revision,
		Payload:          input,
	})
	if err == nil {
		t.Error("mutation succeeded although idempotency record was not persisted")
	}
	stored, ok := app.Get(created.CaseID)
	if !ok {
		t.Fatal("created case disappeared")
	}
	if stored.Revision != created.Revision || stored.SiteName != "原现场" {
		t.Errorf("failed idempotent mutation partially committed: revision=%d site=%q", stored.Revision, stored.SiteName)
	}
	if _, ok := store.GetIdempotency("profile-request"); ok {
		t.Error("failed idempotency write polluted the in-memory replay cache")
	}
}
