package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mural-biocare/internal/application"
	"mural-biocare/internal/domain"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	App *application.Service
	Mux *http.ServeMux
}

func New(app *application.Service) *Server {
	s := &Server{App: app, Mux: http.NewServeMux()}
	s.routes()
	return s
}
func (s *Server) routes() {
	s.Mux.HandleFunc("/healthz", s.health)
	s.Mux.HandleFunc("/api/v1/cases", s.cases)
	s.Mux.HandleFunc("/api/v1/cases/", s.caseRoute)
}
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
func decode(r *http.Request, v any) error {
	defer r.Body.Close()
	d := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	d.DisallowUnknownFields()
	return d.Decode(v)
}
func cmd(r *http.Request, p any) application.Command {
	rev, _ := strconv.Atoi(r.Header.Get("X-Expected-Revision"))
	return application.Command{RequestID: r.Header.Get("X-Request-ID"), Actor: r.Header.Get("X-Actor"), ExpectedRevision: rev, Payload: p}
}
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	write(w, 200, map[string]string{"status": "ok"})
}

type createReq struct {
	SiteName                string   `json:"site_name"`
	MuralSection            string   `json:"mural_section"`
	SymptomDescription      string   `json:"symptom_description"`
	OwnerID                 string   `json:"owner_id"`
	BaselineTemperatureC    *float64 `json:"baseline_temperature_c"`
	BaselineHumidityPercent *float64 `json:"baseline_humidity_percent"`
}

func (s *Server) cases(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		write(w, 405, map[string]string{"error": "method"})
		return
	}
	var in createReq
	if err := decode(r, &in); err != nil {
		write(w, 400, map[string]string{"error": err.Error()})
		return
	}
	if in.BaselineTemperatureC == nil || in.BaselineHumidityPercent == nil {
		write(w, 400, map[string]string{"error": "environment readings required"})
		return
	}
	c, err := s.App.Create(in.SiteName, in.MuralSection, in.SymptomDescription, in.OwnerID, *in.BaselineTemperatureC, *in.BaselineHumidityPercent, r.Header.Get("X-Request-ID"))
	if err != nil {
		write(w, 400, map[string]string{"error": err.Error()})
		return
	}
	write(w, 201, c)
}
func (s *Server) caseRoute(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/cases/"), "/")
	if len(parts) == 0 {
		return
	}
	id := parts[0]
	if len(parts) == 1 && r.Method == "GET" {
		c, ok := s.App.Get(id)
		if !ok {
			write(w, 404, nil)
			return
		}
		write(w, 200, caseDetail(c))
		return
	}
	if len(parts) == 2 && parts[1] == "timeline" {
		if r.Method != "GET" {
			write(w, http.StatusMethodNotAllowed, nil)
			return
		}
		write(w, 200, s.App.Timeline(id))
		return
	}
	if len(parts) == 2 && parts[1] == "verify" {
		if r.Method != "GET" {
			write(w, http.StatusMethodNotAllowed, nil)
			return
		}
		report, err := s.App.Verification(id)
		if err != nil {
			write(w, 422, report)
			return
		}
		write(w, 200, report)
		return
	}
	if len(parts) == 2 && parts[1] == "archive-readiness" {
		if r.Method != "GET" {
			write(w, http.StatusMethodNotAllowed, nil)
			return
		}
		result, err := s.App.ArchiveReadiness(id)
		if err != nil {
			write(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		write(w, http.StatusOK, result)
		return
	}
	if len(parts) == 2 && parts[1] == "assessment-diff" {
		if r.Method != "GET" {
			write(w, http.StatusMethodNotAllowed, nil)
			return
		}
		from, errFrom := strconv.Atoi(r.URL.Query().Get("from_version"))
		to, errTo := strconv.Atoi(r.URL.Query().Get("to_version"))
		if errFrom != nil || errTo != nil {
			write(w, http.StatusBadRequest, map[string]string{"error": "from_version and to_version required"})
			return
		}
		result, err := s.App.AssessmentDiff(id, from, to)
		if err != nil {
			status := http.StatusBadRequest
			if errors.Is(err, application.ErrNotFound) {
				status = http.StatusNotFound
			}
			write(w, status, map[string]string{"error": err.Error()})
			return
		}
		write(w, http.StatusOK, result)
		return
	}
	if r.Method != "POST" {
		write(w, 405, nil)
		return
	}
	if strings.TrimSpace(actor(r)) == "" {
		write(w, http.StatusUnauthorized, map[string]string{"error": "X-Actor required"})
		return
	}
	var body map[string]any
	if err := decode(r, &body); err != nil {
		write(w, 400, map[string]string{"error": err.Error()})
		return
	}
	c, err := s.dispatch(id, parts[1], body, r)
	if err != nil {
		code := 422
		if errors.Is(err, domain.ErrInvalid) {
			code = 400
		}
		if errors.Is(err, domain.ErrConflict) {
			code = 409
		}
		if errors.Is(err, domain.ErrForbidden) {
			code = http.StatusConflict
		}
		if errors.Is(err, domain.ErrArchived) {
			code = 409
		}
		if errors.Is(err, application.ErrNotFound) {
			code = http.StatusNotFound
		}
		if errors.Is(err, application.ErrUnauthorized) {
			code = http.StatusForbidden
		}
		if strings.Contains(err.Error(), "request_id conflict") {
			code = http.StatusConflict
		}
		write(w, code, map[string]string{"error": err.Error()})
		return
	}
	write(w, 200, c)
}
func (s *Server) dispatch(id, action string, b map[string]any, r *http.Request) (*domain.TreatmentCase, error) {
	_, ok := s.App.Get(id)
	if !ok {
		return nil, application.ErrNotFound
	}
	co := cmd(r, b)
	if co.ExpectedRevision == 0 {
		if n, ok := b["expected_revision"].(float64); ok {
			co.ExpectedRevision = int(n)
		}
	}
	switch action {
	case "profile-correction":
		if err := rejectUnknown(b, "site_name", "mural_section", "symptom_description", "owner_id", "new_owner_id", "correction_reason", "expected_revision"); err != nil {
			return nil, err
		}
		in := domain.ProfileCorrectionInput{Reason: str(b["correction_reason"])}
		if v, exists := b["site_name"]; exists {
			x, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("%w: site_name", domain.ErrInvalid)
			}
			in.SiteName = &x
		}
		if v, exists := b["mural_section"]; exists {
			x, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("%w: mural_section", domain.ErrInvalid)
			}
			in.MuralSection = &x
		}
		if v, exists := b["symptom_description"]; exists {
			x, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("%w: symptom_description", domain.ErrInvalid)
			}
			in.SymptomDescription = &x
		}
		if v, exists := b["owner_id"]; exists {
			x, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("%w: owner_id", domain.ErrInvalid)
			}
			in.OwnerID = &x
		}
		if v, exists := b["new_owner_id"]; exists {
			if in.OwnerID != nil {
				return nil, fmt.Errorf("%w: duplicate owner correction", domain.ErrInvalid)
			}
			x, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("%w: new_owner_id", domain.ErrInvalid)
			}
			in.OwnerID = &x
		}
		return s.App.CorrectProfile(id, in, co)
	case "baseline", "retest", "environment":
		mt, _ := time.Parse(time.RFC3339, str(b["measured_at"]))
		if mt.IsZero() {
			mt = time.Now()
		}
		return s.App.UpdateBaseline(id, num(b["temperature_c"]), num(b["humidity_percent"]), mt, co)
	case "assessment":
		if err := rejectUnknown(b, "sample_points", "organism_findings", "activity_level", "spread_boundary", "method", "assessor_id", "expected_revision"); err != nil {
			return nil, err
		}
		a := domain.ContaminationAssessment{OrganismFindings: str(b["organism_findings"]), ActivityLevel: str(b["activity_level"]), SpreadBoundary: str(b["spread_boundary"]), Method: str(b["method"]), AssessorID: str(b["assessor_id"])}
		if a.AssessorID == "" {
			a.AssessorID = co.Actor
		}
		if arr, ok := b["sample_points"].([]any); ok {
			for _, v := range arr {
				if m, ok := v.(map[string]any); ok {
					if err := rejectUnknown(m, "id", "location", "result", "collected_at"); err != nil {
						return nil, err
					}
					sp := domain.SamplePoint{ID: str(m["id"]), Location: str(m["location"]), Result: str(m["result"])}
					if ts := str(m["collected_at"]); ts != "" {
						sp.CollectedAt, _ = time.Parse(time.RFC3339, ts)
					}
					a.SamplePoints = append(a.SamplePoints, sp)
				}
			}
		}
		return s.App.Assessment(id, a, co)
	case "plan":
		return s.App.Plan(id, domain.TreatmentPlan{MaterialName: str(b["material_name"]), CompatibilityBasis: str(b["compatibility_basis"]), ApplicationParameters: str(b["application_parameters"]), ProtectionMeasures: str(b["protection_measures"]), RollbackConditions: str(b["rollback_conditions"]), RequiredObservationDays: int(num(b["required_observation_days"]))}, co)
	case "review":
		if err := rejectUnknown(b, "version", "decision", "comment", "review_items", "expected_revision"); err != nil {
			return nil, err
		}
		if int(num(b["version"])) <= 0 {
			return nil, fmt.Errorf("%w: plan version required", domain.ErrInvalid)
		}
		items, err := reviewItemsFrom(b["review_items"])
		if err != nil {
			return nil, err
		}
		return s.App.ReviewChecklist(id, int(num(b["version"])), str(b["decision"]), str(b["comment"]), items, co)
	case "pilot":
		if err := rejectUnknown(b, "before_activity", "after_activity", "before_color_delta", "after_color_delta", "color_threshold", "observation_days", "notes", "observations", "expected_revision"); err != nil {
			return nil, err
		}
		observations, err := observationsFrom(b["observations"])
		if err != nil {
			return nil, err
		}
		return s.App.Pilot(id, domain.PilotTrial{BeforeActivity: num(b["before_activity"]), AfterActivity: num(b["after_activity"]), BeforeColorDelta: num(b["before_color_delta"]), AfterColorDelta: num(b["after_color_delta"]), ColorThreshold: num(b["color_threshold"]), ObservationDays: int(num(b["observation_days"])), Notes: str(b["notes"]), Observations: observations}, co)
	case "start":
		if err := rejectUnknown(b, "planned_checkpoints", "expected_revision"); err != nil {
			return nil, err
		}
		plan, err := plannedCheckpointsFrom(b["planned_checkpoints"])
		if err != nil {
			return nil, err
		}
		return s.App.StartWithPlan(id, plan, co)
	case "checkpoint":
		if raw, ok := b["checkpoints"].([]any); ok {
			checkpoints := make([]domain.ExecutionCheckpoint, 0, len(raw))
			for _, item := range raw {
				m, ok := item.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("%w: checkpoint item", domain.ErrInvalid)
				}
				cp, err := checkpointFromMap(m, co.Actor)
				if err != nil {
					return nil, err
				}
				checkpoints = append(checkpoints, cp)
			}
			return s.App.Checkpoints(id, checkpoints, co)
		}
		observed := str(b["observed_value"])
		if d := str(b["deviation_description"]); d != "" {
			observed = d
		}
		cp := domain.ExecutionCheckpoint{CheckpointID: str(b["checkpoint_id"]), Sequence: int(num(b["sequence"])), Phase: str(b["phase"]), ExpectedCondition: str(b["expected_condition"]), ObservedValue: observed, Result: str(b["result"]), DeviationID: str(b["deviation_id"]), DeviationType: str(b["deviation_type"]), RelatedDeviationID: str(b["related_deviation_id"]), RectificationItemID: str(b["rectification_item_id"]), Unit: str(b["unit"]), RecordedBy: co.Actor}
		return s.App.Checkpoint(id, cp, co)
	case "resolve":
		refs := stringsFrom(b["evidence_refs"])
		if len(refs) == 0 {
			refs = stringsFrom(b["evidence"])
		}
		action := str(b["corrective_action"])
		if action == "" {
			action = str(b["evidence"])
		}
		if action == "" && len(refs) > 0 {
			action = "依据所附证据完成纠正"
		}
		return s.App.ResolveWithEvidence(id, str(b["deviation_id"]), action, refs, str(b["reviewer_id"]), co)
	case "complete":
		return s.App.Complete(id, co)
	case "outcome":
		o := domain.OutcomeVerification{PostActivity: num(b["post_activity"]), ColorDelta: num(b["color_delta"]), SurfaceStability: num(b["surface_stability"]), ActivityThreshold: num(b["activity_threshold"]), ColorThreshold: num(b["color_threshold"]), StabilityThreshold: num(b["stability_threshold"]), ObservationDays: int(num(b["observation_days"])), Rectification: str(b["rectification"])}
		if ts := str(b["measured_at"]); ts != "" {
			o.MeasuredAt, _ = time.Parse(time.RFC3339, ts)
		}
		if o.MeasuredAt.IsZero() {
			o.MeasuredAt = time.Now()
		}
		return s.App.Outcome(id, o, co)
	case "archive":
		return s.App.Archive(id, co)
	}
	return nil, application.ErrNotFound
}
func checkpointFromMap(b map[string]any, actor string) (domain.ExecutionCheckpoint, error) {
	observed := str(b["observed_value"])
	if d := str(b["deviation_description"]); d != "" {
		observed = d
	}
	cp := domain.ExecutionCheckpoint{CheckpointID: str(b["checkpoint_id"]), Sequence: int(num(b["sequence"])), Phase: str(b["phase"]), ExpectedCondition: str(b["expected_condition"]), ObservedValue: observed, Result: str(b["result"]), DeviationID: str(b["deviation_id"]), DeviationType: str(b["deviation_type"]), RelatedDeviationID: str(b["related_deviation_id"]), RectificationItemID: str(b["rectification_item_id"]), Unit: str(b["unit"]), RecordedBy: actor}
	if ts := str(b["recorded_at"]); ts != "" {
		var err error
		cp.RecordedAt, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			return cp, fmt.Errorf("%w: recorded_at", domain.ErrInvalid)
		}
	}
	return cp, nil
}
func reviewItemsFrom(v any) ([]domain.PlanReviewItem, error) {
	raw, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: review_items required", domain.ErrInvalid)
	}
	items := make([]domain.PlanReviewItem, 0, len(raw))
	for _, entry := range raw {
		m, ok := entry.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%w: review item", domain.ErrInvalid)
		}
		if err := rejectUnknown(m, "item", "result", "comment", "modification_comment"); err != nil {
			return nil, err
		}
		comment := str(m["comment"])
		if comment == "" {
			comment = str(m["modification_comment"])
		}
		items = append(items, domain.PlanReviewItem{Item: str(m["item"]), Result: str(m["result"]), Comment: comment})
	}
	return items, nil
}
func observationsFrom(v any) ([]domain.PilotObservation, error) {
	if v == nil {
		return nil, nil
	}
	raw, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: observations", domain.ErrInvalid)
	}
	items := make([]domain.PilotObservation, 0, len(raw))
	for _, entry := range raw {
		m, ok := entry.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%w: observation item", domain.ErrInvalid)
		}
		if err := rejectUnknown(m, "observation_day", "activity", "color_delta"); err != nil {
			return nil, err
		}
		items = append(items, domain.PilotObservation{ObservationDay: int(num(m["observation_day"])), Activity: num(m["activity"]), ColorDelta: num(m["color_delta"])})
	}
	return items, nil
}
func plannedCheckpointsFrom(v any) ([]domain.PlannedCheckpoint, error) {
	raw, ok := v.([]any)
	if !ok || len(raw) == 0 {
		return nil, fmt.Errorf("%w: planned_checkpoints required", domain.ErrInvalid)
	}
	items := make([]domain.PlannedCheckpoint, 0, len(raw))
	for _, entry := range raw {
		m, ok := entry.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%w: planned checkpoint", domain.ErrInvalid)
		}
		if err := rejectUnknown(m, "checkpoint_id", "sequence", "phase", "expected_condition", "unit"); err != nil {
			return nil, err
		}
		items = append(items, domain.PlannedCheckpoint{CheckpointID: str(m["checkpoint_id"]), Sequence: int(num(m["sequence"])), Phase: str(m["phase"]), ExpectedCondition: str(m["expected_condition"]), Unit: str(m["unit"])})
	}
	return items, nil
}
func rejectUnknown(m map[string]any, allowed ...string) error {
	set := make(map[string]bool, len(allowed))
	for _, key := range allowed {
		set[key] = true
	}
	for key := range m {
		if !set[key] {
			return fmt.Errorf("%w: unknown field %s", domain.ErrInvalid, key)
		}
	}
	return nil
}
func stringsFrom(v any) []string {
	if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
		return []string{s}
	}
	a, ok := v.([]any)
	if !ok {
		return nil
	}
	r := make([]string, 0, len(a))
	for _, item := range a {
		if s, ok := item.(string); ok {
			r = append(r, s)
		} else {
			return nil
		}
	}
	return r
}
func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
func num(v any) float64 {
	if n, ok := v.(float64); ok {
		return n
	}
	return 0
}
func caseDetail(c *domain.TreatmentCase) map[string]any {
	b, _ := json.Marshal(c)
	var out map[string]any
	_ = json.Unmarshal(b, &out)
	out["baseline"] = map[string]any{"temperature_c": c.BaselineTemperatureC, "humidity_percent": c.BaselineHumidityPercent, "status": c.BaselineStatus, "retest_count": c.BaselineRetestCount, "latest_anomaly_reason": c.LatestBaselineAnomalyReason, "trend": c.BaselineMeasurements, "unlocked_at": c.BaselineUnlockedAt}
	chains := map[string][]domain.Deviation{}
	for _, d := range c.Deviations {
		root := d.RootDeviationID
		if root == "" {
			root = d.ID
		}
		chains[root] = append(chains[root], d)
	}
	out["deviation_recurrence_chains"] = chains
	return out
}
