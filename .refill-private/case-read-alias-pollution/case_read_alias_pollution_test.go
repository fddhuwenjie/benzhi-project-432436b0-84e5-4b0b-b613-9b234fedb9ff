package case_read_alias_pollution_test

import (
	"mural-biocare/internal/application"
	"mural-biocare/internal/domain"
	"mural-biocare/internal/persistence"
	"testing"
	"time"
)

func TestCaseReadDoesNotAliasStoredAssessment(t *testing.T) {
	store, err := persistence.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	app := application.New(store)
	c, err := app.Create("永宁寺", "东壁", "白色菌斑", "owner-1", 20, 52, "create-alias-case")
	if err != nil {
		t.Fatal(err)
	}
	assessment := domain.ContaminationAssessment{
		SamplePoints: []domain.SamplePoint{
			{ID: "sample-1", Location: "东壁上部", Result: "positive", CollectedAt: c.CreatedAt},
			{ID: "sample-2", Location: "东壁下部", Result: "negative", CollectedAt: c.CreatedAt.Add(time.Nanosecond)},
		},
		OrganismFindings: "曲霉属",
		ActivityLevel:    "HIGH",
		SpreadBoundary:   "东壁局部",
		Method:           "culture",
		AssessorID:       "assessor-1",
	}
	c, err = app.Assessment(c.CaseID, assessment, application.Command{
		Actor:            "assessor-1",
		ExpectedRevision: c.Revision,
		Payload:          assessment,
	})
	if err != nil {
		t.Fatal(err)
	}

	readModel, ok := app.Get(c.CaseID)
	if !ok || readModel.Assessment == nil {
		t.Fatal("assessment missing from case read")
	}
	readModel.Assessment.OrganismFindings = "调用方篡改"
	readModel.Assessment.SamplePoints[0].Result = "tampered"
	readModel.Assessment.SamplePoints = append(readModel.Assessment.SamplePoints, domain.SamplePoint{ID: "injected"})

	stored, ok := app.Get(c.CaseID)
	if !ok || stored.Assessment == nil {
		t.Fatal("stored assessment missing")
	}
	if stored.Assessment.OrganismFindings != "曲霉属" || stored.Assessment.SamplePoints[0].Result != "positive" || len(stored.Assessment.SamplePoints) != 2 {
		t.Fatalf("TestCaseReadDoesNotAliasStoredAssessment: read result polluted stored assessment: %+v", stored.Assessment)
	}
}
