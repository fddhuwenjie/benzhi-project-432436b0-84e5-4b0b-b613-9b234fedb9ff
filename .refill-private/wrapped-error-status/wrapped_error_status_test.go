package wrapped_error_status_test

import (
	"bytes"
	"encoding/json"
	"mural-biocare/internal/application"
	"mural-biocare/internal/domain"
	"mural-biocare/internal/httpapi"
	"mural-biocare/internal/persistence"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestWrappedApplicationErrorsPreserveHTTPStatus(t *testing.T) {
	store, err := persistence.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	app := application.New(store)
	handler := httpapi.New(app).Handler()
	c, err := app.Create("现场", "A区", "菌斑", "owner", 20, 50, "")
	if err != nil {
		t.Fatal(err)
	}

	conflict := postCorrection(t, handler, c.CaseID, c.Revision+1, map[string]any{
		"site_name":         "修正现场",
		"correction_reason": "更正登记",
		"expected_revision": c.Revision + 1,
	})
	if conflict.Code != http.StatusConflict {
		t.Fatalf("wrapped revision conflict must be HTTP 409, got %d: %s", conflict.Code, conflict.Body.String())
	}

	invalid := postCorrection(t, handler, c.CaseID, c.Revision, map[string]any{
		"site_name":         42,
		"correction_reason": "更正登记",
		"expected_revision": c.Revision,
	})
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("wrapped invalid command must be HTTP 400, got %d: %s", invalid.Code, invalid.Body.String())
	}

	assessment := domain.ContaminationAssessment{
		SamplePoints:     []domain.SamplePoint{{ID: "s1", Location: "A区", Result: "positive", CollectedAt: time.Now()}},
		OrganismFindings: "真菌",
		ActivityLevel:    "HIGH",
		SpreadBoundary:   "A区",
		Method:           "culture",
		AssessorID:       "assessor",
	}
	c, err = app.Assessment(c.CaseID, assessment, application.Command{Actor: "assessor", ExpectedRevision: c.Revision, Payload: assessment})
	if err != nil {
		t.Fatal(err)
	}
	forbidden := postCorrection(t, handler, c.CaseID, c.Revision, map[string]any{
		"site_name":         "状态迁移后修改",
		"correction_reason": "不应允许",
		"expected_revision": c.Revision,
	})
	if forbidden.Code != http.StatusConflict {
		t.Fatalf("wrapped forbidden transition must be HTTP 409, got %d: %s", forbidden.Code, forbidden.Body.String())
	}

	unauthorized := postCorrectionAs(t, handler, c.CaseID, "another-actor", c.Revision, map[string]any{
		"site_name":         "越权修改",
		"correction_reason": "无权限",
		"expected_revision": c.Revision,
	})
	if unauthorized.Code != http.StatusForbidden {
		t.Fatalf("direct authorization rejection must remain HTTP 403, got %d: %s", unauthorized.Code, unauthorized.Body.String())
	}
}

func postCorrection(t *testing.T, handler http.Handler, caseID string, revision int, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	return postCorrectionAs(t, handler, caseID, "owner", revision, body)
}

func postCorrectionAs(t *testing.T, handler http.Handler, caseID, actor string, revision int, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cases/"+caseID+"/profile-correction", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Actor", actor)
	req.Header.Set("X-Expected-Revision", strconv.Itoa(revision))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}
