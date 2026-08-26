package domain

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type Status string

const (
	StatusDraft              Status = "DRAFT"
	StatusAssessed           Status = "ASSESSED"
	StatusPlanApproved       Status = "PLAN_APPROVED"
	StatusPilotPassed        Status = "PILOT_PASSED"
	StatusInProgress         Status = "IN_PROGRESS"
	StatusPaused             Status = "PAUSED"
	StatusTreatmentCompleted Status = "TREATMENT_COMPLETED"
	StatusOutcomeVerified    Status = "OUTCOME_VERIFIED"
	StatusArchived           Status = "ARCHIVED"
	StatusPlanRevision       Status = "PLAN_REVISION_REQUIRED"
	StatusPilotRevision      Status = "PILOT_REVISION_REQUIRED"
)

const (
	BaselineNormal         = "NORMAL"
	BaselineRetestRequired = "RETEST_REQUIRED"
)

type TreatmentCase struct {
	CaseID                      string                    `json:"case_id"`
	SiteName                    string                    `json:"site_name"`
	MuralSection                string                    `json:"mural_section"`
	SymptomDescription          string                    `json:"symptom_description"`
	BaselineTemperatureC        float64                   `json:"baseline_temperature_c"`
	BaselineHumidityPercent     float64                   `json:"baseline_humidity_percent"`
	BaselineStatus              string                    `json:"baseline_status,omitempty"`
	BaselineAnomalyReason       string                    `json:"baseline_anomaly_reason,omitempty"`
	LatestBaselineAnomalyReason string                    `json:"latest_baseline_anomaly_reason,omitempty"`
	BaselineMeasurements        []BaselineMeasurement     `json:"baseline_measurements,omitempty"`
	BaselineRetestCount         int                       `json:"baseline_retest_count"`
	BaselineUnlockedAt          *time.Time                `json:"baseline_unlocked_at,omitempty"`
	OwnerID                     string                    `json:"owner_id"`
	ExecutionResponsible        string                    `json:"execution_responsible,omitempty"`
	Status                      Status                    `json:"status"`
	Revision                    int                       `json:"revision"`
	CreatedAt                   time.Time                 `json:"created_at"`
	ArchivedAt                  *time.Time                `json:"archived_at,omitempty"`
	Assessment                  *ContaminationAssessment  `json:"assessment,omitempty"`
	AssessmentHistory           []ContaminationAssessment `json:"assessment_history,omitempty"`
	Plan                        *TreatmentPlan            `json:"plan,omitempty"`
	PlanHistory                 []TreatmentPlan           `json:"plan_history,omitempty"`
	Pilot                       *PilotTrial               `json:"pilot,omitempty"`
	PilotHistory                []PilotTrial              `json:"pilot_history,omitempty"`
	Checkpoints                 []ExecutionCheckpoint     `json:"checkpoints,omitempty"`
	CheckpointSummary           []string                  `json:"checkpoint_summary,omitempty"`
	CheckpointCompletionSummary []CheckpointCompletion    `json:"checkpoint_completion_summary,omitempty"`
	Outcome                     *OutcomeVerification      `json:"outcome,omitempty"`
	OutcomeHistory              []OutcomeVerification     `json:"outcome_history,omitempty"`
	Deviations                  []Deviation               `json:"deviations,omitempty"`
	ProfileCorrections          []ProfileCorrection       `json:"profile_corrections,omitempty"`
	AssessmentDiffs             []AssessmentDiff          `json:"assessment_diffs,omitempty"`
	PlannedCheckpoints          []PlannedCheckpoint       `json:"planned_checkpoints,omitempty"`
	RectificationItems          []RectificationItem       `json:"rectification_items,omitempty"`
	Evidence                    *EvidenceManifest         `json:"evidence,omitempty"`
}

type FieldChange struct {
	Field    string `json:"field"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}

type ProfileCorrection struct {
	Revision      int           `json:"revision"`
	Reason        string        `json:"correction_reason"`
	Actor         string        `json:"actor"`
	PreviousOwner string        `json:"previous_owner"`
	NewOwner      string        `json:"new_owner"`
	Changes       []FieldChange `json:"changes"`
	CorrectedAt   time.Time     `json:"corrected_at"`
}

type BaselineMeasurement struct {
	TemperatureC    float64   `json:"temperature_c"`
	HumidityPercent float64   `json:"humidity_percent"`
	MeasuredAt      time.Time `json:"measured_at"`
	MeasuredBy      string    `json:"measured_by"`
	Revision        int       `json:"revision"`
	Status          string    `json:"status"`
	AnomalyReason   string    `json:"anomaly_reason,omitempty"`
}

type ContaminationAssessment struct {
	AssessmentID     string        `json:"assessment_id"`
	CaseID           string        `json:"case_id"`
	SamplePoints     []SamplePoint `json:"sample_points"`
	OrganismFindings string        `json:"organism_findings"`
	ActivityLevel    string        `json:"activity_level"`
	SpreadBoundary   string        `json:"spread_boundary"`
	Method           string        `json:"method"`
	AssessorID       string        `json:"assessor_id"`
	AssessedAt       time.Time     `json:"assessed_at"`
	Version          int           `json:"version"`
}
type SamplePointChange struct {
	ID     string       `json:"id"`
	Before *SamplePoint `json:"before,omitempty"`
	After  *SamplePoint `json:"after,omitempty"`
}
type AssessmentDiff struct {
	FromVersion   int                 `json:"from_version"`
	ToVersion     int                 `json:"to_version"`
	Added         []SamplePoint       `json:"added_sample_points,omitempty"`
	Removed       []SamplePoint       `json:"removed_sample_points,omitempty"`
	Changed       []SamplePointChange `json:"changed_sample_points,omitempty"`
	ChangedFields []FieldChange       `json:"changed_fields,omitempty"`
}
type SamplePoint struct {
	ID          string    `json:"id"`
	Location    string    `json:"location"`
	Result      string    `json:"result"`
	CollectedAt time.Time `json:"collected_at"`
}
type TreatmentPlan struct {
	PlanID                  string           `json:"plan_id"`
	CaseID                  string           `json:"case_id"`
	MaterialName            string           `json:"material_name"`
	CompatibilityBasis      string           `json:"compatibility_basis"`
	ApplicationParameters   string           `json:"application_parameters"`
	ProtectionMeasures      string           `json:"protection_measures"`
	RollbackConditions      string           `json:"rollback_conditions"`
	AuthorID                string           `json:"author_id"`
	ReviewerID              string           `json:"reviewer_id,omitempty"`
	ReviewDecision          string           `json:"review_decision,omitempty"`
	ReviewComment           string           `json:"review_comment,omitempty"`
	Version                 int              `json:"version"`
	RequiredObservationDays int              `json:"required_observation_days,omitempty"`
	PreviousReviewerID      string           `json:"previous_reviewer_id,omitempty"`
	ChangedFields           []string         `json:"changed_fields,omitempty"`
	ReviewItems             []PlanReviewItem `json:"review_items,omitempty"`
}
type PlanReviewItem struct {
	Item    string `json:"item"`
	Result  string `json:"result"`
	Comment string `json:"comment,omitempty"`
}
type PilotObservation struct {
	ObservationDay int     `json:"observation_day"`
	Activity       float64 `json:"activity"`
	ColorDelta     float64 `json:"color_delta"`
}
type PilotTrial struct {
	BeforeActivity        float64            `json:"before_activity"`
	AfterActivity         float64            `json:"after_activity"`
	BeforeColorDelta      float64            `json:"before_color_delta"`
	AfterColorDelta       float64            `json:"after_color_delta"`
	ColorThreshold        float64            `json:"color_threshold"`
	ObservationDays       int                `json:"observation_days"`
	Passed                bool               `json:"passed"`
	Notes                 string             `json:"notes,omitempty"`
	ActivityGatePassed    bool               `json:"activity_gate_passed"`
	ColorGatePassed       bool               `json:"color_gate_passed"`
	ObservationGatePassed bool               `json:"observation_gate_passed"`
	FailureReasons        []string           `json:"failure_reasons,omitempty"`
	PlanVersion           int                `json:"plan_version,omitempty"`
	Round                 int                `json:"round"`
	RecordedAt            time.Time          `json:"recorded_at"`
	Observations          []PilotObservation `json:"observations,omitempty"`
}
type PlannedCheckpoint struct {
	CheckpointID      string `json:"checkpoint_id"`
	Sequence          int    `json:"sequence"`
	Phase             string `json:"phase"`
	ExpectedCondition string `json:"expected_condition"`
	Unit              string `json:"unit,omitempty"`
}
type CheckpointCompletion struct {
	CheckpointID      string `json:"checkpoint_id"`
	Sequence          int    `json:"sequence"`
	Phase             string `json:"phase"`
	ExpectedCondition string `json:"expected_condition"`
	ObservedValue     string `json:"observed_value"`
	Unit              string `json:"unit,omitempty"`
	Result            string `json:"result"`
}
type ExecutionCheckpoint struct {
	CheckpointID        string    `json:"checkpoint_id"`
	CaseID              string    `json:"case_id"`
	Phase               string    `json:"phase"`
	Sequence            int       `json:"sequence"`
	ExpectedCondition   string    `json:"expected_condition"`
	ObservedValue       string    `json:"observed_value"`
	Result              string    `json:"result"`
	DeviationID         string    `json:"deviation_id,omitempty"`
	EvidenceRefs        []string  `json:"evidence_refs,omitempty"`
	RecordedBy          string    `json:"recorded_by"`
	RecordedAt          time.Time `json:"recorded_at"`
	Unit                string    `json:"unit,omitempty"`
	DeviationType       string    `json:"deviation_type,omitempty"`
	RelatedDeviationID  string    `json:"related_deviation_id,omitempty"`
	RectificationItemID string    `json:"rectification_item_id,omitempty"`
}
type Deviation struct {
	ID                 string     `json:"id"`
	Description        string     `json:"description"`
	CorrectiveEvidence string     `json:"corrective_evidence,omitempty"`
	ReportedBy         string     `json:"reported_by"`
	Resolved           bool       `json:"resolved"`
	CreatedAt          *time.Time `json:"created_at"`
	ResolvedAt         *time.Time `json:"resolved_at,omitempty"`
	Type               string     `json:"type,omitempty"`
	Severity           string     `json:"severity"`
	EvidenceRefs       []string   `json:"evidence_refs,omitempty"`
	CorrectiveAction   string     `json:"corrective_action,omitempty"`
	ReviewerID         string     `json:"reviewer_id,omitempty"`
	ResolvedBy         string     `json:"resolved_by,omitempty"`
	RelatedDeviationID string     `json:"related_deviation_id,omitempty"`
	RootDeviationID    string     `json:"root_deviation_id,omitempty"`
	RecurrenceCount    int        `json:"recurrence_count"`
	EscalationBasis    string     `json:"escalation_basis,omitempty"`
}
type OutcomeVerification struct {
	PostActivity       float64            `json:"post_activity"`
	ColorDelta         float64            `json:"color_delta"`
	SurfaceStability   float64            `json:"surface_stability"`
	ObservationDays    int                `json:"observation_days"`
	ActivityThreshold  float64            `json:"activity_threshold"`
	ColorThreshold     float64            `json:"color_threshold"`
	StabilityThreshold float64            `json:"stability_threshold"`
	VerifiedBy         string             `json:"verified_by"`
	Passed             bool               `json:"passed"`
	Rectification      string             `json:"rectification,omitempty"`
	VerifiedAt         time.Time          `json:"verified_at"`
	Round              int                `json:"round"`
	MeasuredAt         time.Time          `json:"measured_at"`
	FailureItems       []string           `json:"failure_items,omitempty"`
	ThresholdSnapshot  map[string]float64 `json:"threshold_snapshot,omitempty"`
}
type RectificationItem struct {
	ID                   string  `json:"rectification_item_id"`
	OutcomeRound         int     `json:"outcome_round"`
	Metric               string  `json:"metric"`
	FailureValue         float64 `json:"failure_value"`
	TargetThreshold      float64 `json:"target_threshold"`
	Requirement          string  `json:"requirement"`
	Status               string  `json:"status"`
	CheckpointID         string  `json:"checkpoint_id,omitempty"`
	ClosedByOutcomeRound int     `json:"closed_by_outcome_round,omitempty"`
}
type EvidenceManifest struct {
	ManifestID     string              `json:"manifest_id"`
	CaseID         string              `json:"case_id"`
	CaseRevision   int                 `json:"case_revision"`
	EvidenceItems  []string            `json:"evidence_items"`
	EventChainHead string              `json:"event_chain_head"`
	ContentDigest  string              `json:"content_digest"`
	SealedBy       string              `json:"sealed_by"`
	SealedAt       time.Time           `json:"sealed_at"`
	EvidenceIndex  map[string][]string `json:"evidence_index"`
}

var ErrInvalid = errors.New("invalid command")
var ErrConflict = errors.New("revision conflict")
var ErrForbidden = errors.New("forbidden state transition")
var ErrArchived = errors.New("case archived")

func NewCase(id, site, section, symptom, owner string, temp, humidity float64, now time.Time) (*TreatmentCase, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(site) == "" || strings.TrimSpace(section) == "" || strings.TrimSpace(symptom) == "" || strings.TrimSpace(owner) == "" {
		return nil, fmt.Errorf("%w: required fields", ErrInvalid)
	}
	if !finite(temp) || !finite(humidity) || temp < -50 || temp > 80 || humidity < 0 || humidity > 100 || !precisionOK(temp) || !precisionOK(humidity) {
		return nil, fmt.Errorf("%w: environment range", ErrInvalid)
	}
	status, reason := baselineClass(temp, humidity)
	return &TreatmentCase{CaseID: id, SiteName: site, MuralSection: section, SymptomDescription: symptom, OwnerID: owner, BaselineTemperatureC: temp, BaselineHumidityPercent: humidity, BaselineStatus: status, BaselineAnomalyReason: reason, LatestBaselineAnomalyReason: reason, Status: StatusDraft, Revision: 1, CreatedAt: now, BaselineMeasurements: []BaselineMeasurement{{TemperatureC: temp, HumidityPercent: humidity, MeasuredAt: now, MeasuredBy: owner, Revision: 1, Status: status, AnomalyReason: reason}}}, nil
}

func finite(v float64) bool      { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func precisionOK(v float64) bool { return math.Abs(v*100-math.Round(v*100)) < 1e-9 }
func baselineClass(temp, humidity float64) (string, string) {
	if temp < 5 || temp > 35 || humidity < 20 || humidity > 80 || (temp > 30 && humidity > 70) {
		return "RETEST_REQUIRED", "温湿度组合超出壁画现场保护基线"
	}
	return "NORMAL", ""
}
func (c *TreatmentCase) UpdateBaseline(temp, humidity float64, measuredAt time.Time, by string, rev int) error {
	return c.UpdateBaselineWithInterval(temp, humidity, measuredAt, by, rev, 24*time.Hour)
}
func (c *TreatmentCase) UpdateBaselineWithInterval(temp, humidity float64, measuredAt time.Time, by string, rev int, interval time.Duration) error {
	if err := c.check(rev); err != nil {
		return err
	}
	if by == "" || measuredAt.IsZero() || !finite(temp) || !finite(humidity) || temp < -50 || temp > 80 || humidity < 0 || humidity > 100 || !precisionOK(temp) || !precisionOK(humidity) {
		return fmt.Errorf("%w: invalid baseline", ErrInvalid)
	}
	if len(c.BaselineMeasurements) > 0 && !measuredAt.After(c.BaselineMeasurements[len(c.BaselineMeasurements)-1].MeasuredAt) {
		return fmt.Errorf("%w: measured_at must increase", ErrInvalid)
	}
	readingStatus, reason := baselineClass(temp, humidity)
	c.BaselineTemperatureC, c.BaselineHumidityPercent = temp, humidity
	c.BaselineRetestCount++
	c.BaselineMeasurements = append(c.BaselineMeasurements, BaselineMeasurement{TemperatureC: temp, HumidityPercent: humidity, MeasuredAt: measuredAt, MeasuredBy: by, Revision: c.Revision + 1, Status: readingStatus, AnomalyReason: reason})
	if readingStatus == BaselineRetestRequired {
		c.BaselineStatus, c.BaselineAnomalyReason, c.BaselineUnlockedAt = readingStatus, reason, nil
		c.LatestBaselineAnomalyReason = reason
	} else if c.BaselineStatus == BaselineRetestRequired {
		n := len(c.BaselineMeasurements)
		if n >= 3 && c.BaselineMeasurements[n-2].Status == BaselineNormal && measuredAt.Sub(c.BaselineMeasurements[n-2].MeasuredAt) >= interval {
			c.BaselineStatus, c.BaselineAnomalyReason = BaselineNormal, ""
			unlocked := measuredAt
			c.BaselineUnlockedAt = &unlocked
		}
	} else {
		c.BaselineStatus, c.BaselineAnomalyReason = BaselineNormal, ""
	}
	c.bump()
	return nil
}

func (c *TreatmentCase) check(rev int) error {
	if c.Status == StatusArchived {
		return ErrArchived
	}
	if rev != c.Revision {
		return fmt.Errorf("%w: expected %d got %d", ErrConflict, rev, c.Revision)
	}
	return nil
}
func (c *TreatmentCase) bump() { c.Revision++ }

func (c *TreatmentCase) SubmitAssessment(a ContaminationAssessment, rev int) error {
	if err := c.check(rev); err != nil {
		return err
	}
	if c.Status != StatusDraft && c.Status != StatusAssessed && c.Status != StatusPlanRevision && c.Status != StatusPilotRevision {
		return fmt.Errorf("%w: assessment", ErrForbidden)
	}
	if c.BaselineStatus == "RETEST_REQUIRED" {
		return fmt.Errorf("%w: baseline retest required: %s", ErrInvalid, c.BaselineAnomalyReason)
	}
	if len(a.SamplePoints) == 0 || strings.TrimSpace(a.OrganismFindings) == "" || strings.TrimSpace(a.ActivityLevel) == "" || strings.TrimSpace(a.SpreadBoundary) == "" || strings.TrimSpace(a.Method) == "" || strings.TrimSpace(a.AssessorID) == "" {
		return fmt.Errorf("%w: incomplete assessment", ErrInvalid)
	}
	seen := map[string]bool{}
	positive := false
	negative := true
	lowConclusion := false
	now := time.Now()
	var previous time.Time
	for _, p := range a.SamplePoints {
		if strings.TrimSpace(p.ID) == "" || strings.TrimSpace(p.Location) == "" || strings.TrimSpace(p.Result) == "" || p.CollectedAt.IsZero() || seen[p.ID] {
			return fmt.Errorf("%w: invalid sample point", ErrInvalid)
		}
		if p.CollectedAt.Before(c.CreatedAt) || p.CollectedAt.After(now) {
			return fmt.Errorf("%w: collected_at outside case lifetime", ErrInvalid)
		}
		if !previous.IsZero() && !p.CollectedAt.After(previous) {
			return fmt.Errorf("%w: collected_at must be strictly increasing", ErrInvalid)
		}
		previous = p.CollectedAt
		if strings.TrimSpace(p.Location) != "" && !strings.Contains(p.Location, c.MuralSection) && !strings.Contains(c.MuralSection, p.Location) {
			return fmt.Errorf("%w: sample point outside mural section", ErrInvalid)
		}
		seen[p.ID] = true
		r := strings.ToLower(strings.TrimSpace(p.Result))
		isNegative := r == "negative" || r == "none" || r == "not detected" || strings.Contains(r, "未检出") || strings.Contains(r, "阴性")
		isLow := r == "low" || r == "low_activity" || strings.Contains(r, "低活性")
		if isLow {
			lowConclusion = true
		}
		if !isNegative && (isLow || strings.Contains(r, "positive") || strings.Contains(r, "detected") || strings.Contains(r, "阳性") || strings.Contains(r, "检出")) {
			positive = true
			negative = false
		}
	}
	a.ActivityLevel = NormalizeActivity(a.ActivityLevel)
	if a.ActivityLevel == "NEGATIVE" {
		a.ActivityLevel = "NONE"
	}
	if a.ActivityLevel != "NONE" && a.ActivityLevel != "LOW" && a.ActivityLevel != "MEDIUM" && a.ActivityLevel != "HIGH" {
		return fmt.Errorf("%w: unsupported activity_level", ErrInvalid)
	}
	findings := strings.ToLower(strings.TrimSpace(a.OrganismFindings))
	boundary := strings.ToLower(strings.TrimSpace(a.SpreadBoundary))
	findingsNegative := findings == "none" || findings == "not detected" || strings.Contains(findings, "未检出")
	boundaryNone := boundary == "none" || boundary == "n/a" || boundary == "无"
	if positive && (a.ActivityLevel == "NONE" || boundaryNone) {
		return fmt.Errorf("%w: positive samples require activity and spread boundary", ErrInvalid)
	}
	if lowConclusion && a.ActivityLevel != "LOW" {
		return fmt.Errorf("%w: low-activity result requires LOW activity_level", ErrInvalid)
	}
	if negative && (!findingsNegative || a.ActivityLevel != "NONE" || !boundaryNone) {
		return fmt.Errorf("%w: non-detected samples require NONE activity and empty spread", ErrInvalid)
	}
	if findingsNegative && (positive || a.ActivityLevel != "NONE") {
		return fmt.Errorf("%w: organism findings conflict with samples", ErrInvalid)
	}
	a.CaseID = c.CaseID
	if a.AssessmentID == "" {
		a.AssessmentID = fmt.Sprintf("asm-%d", time.Now().UnixNano())
	}
	a.Version = 1
	if c.Assessment != nil {
		a.Version = c.Assessment.Version + 1
	}
	a.AssessedAt = now
	if c.Assessment != nil {
		c.AssessmentDiffs = append(c.AssessmentDiffs, BuildAssessmentDiff(*c.Assessment, a))
		c.AssessmentHistory = append(c.AssessmentHistory, *c.Assessment)
	}
	c.Assessment = &a
	c.Status = StatusAssessed
	c.bump()
	return nil
}

func (c *TreatmentCase) SubmitPlan(p TreatmentPlan, rev int) error {
	if err := c.check(rev); err != nil {
		return err
	}
	if c.Status != StatusAssessed && c.Status != StatusPlanRevision && c.Status != StatusPilotRevision {
		return fmt.Errorf("%w: plan", ErrForbidden)
	}
	if strings.TrimSpace(p.MaterialName) == "" || strings.TrimSpace(p.CompatibilityBasis) == "" || strings.TrimSpace(p.ApplicationParameters) == "" || strings.TrimSpace(p.ProtectionMeasures) == "" || strings.TrimSpace(p.RollbackConditions) == "" || strings.TrimSpace(p.AuthorID) == "" || p.RequiredObservationDays <= 0 || p.RequiredObservationDays > 365 {
		return fmt.Errorf("%w: incomplete plan", ErrInvalid)
	}
	if p.Version < 1 {
		p.Version = 1
	}
	if c.Plan != nil {
		p.Version = c.Plan.Version + 1
		p.PreviousReviewerID = c.Plan.ReviewerID
		p.ChangedFields = planChanges(*c.Plan, p)
	} else {
		p.ChangedFields = []string{"material_name", "compatibility_basis", "application_parameters", "protection_measures", "rollback_conditions", "required_observation_days"}
	}
	p.ReviewerID, p.ReviewDecision, p.ReviewComment = "", "", ""
	p.CaseID = c.CaseID
	if p.PlanID == "" {
		p.PlanID = fmt.Sprintf("plan-%d", time.Now().UnixNano())
	}
	if c.Plan != nil {
		c.PlanHistory = append(c.PlanHistory, *c.Plan)
	}
	c.Plan = &p
	c.Status = StatusAssessed
	c.bump()
	return nil
}
func planChanges(old, next TreatmentPlan) []string {
	changes := []string{}
	if old.MaterialName != next.MaterialName {
		changes = append(changes, "material_name")
	}
	if old.CompatibilityBasis != next.CompatibilityBasis {
		changes = append(changes, "compatibility_basis")
	}
	if old.ApplicationParameters != next.ApplicationParameters {
		changes = append(changes, "application_parameters")
	}
	if old.ProtectionMeasures != next.ProtectionMeasures {
		changes = append(changes, "protection_measures")
	}
	if old.RollbackConditions != next.RollbackConditions {
		changes = append(changes, "rollback_conditions")
	}
	if old.RequiredObservationDays != next.RequiredObservationDays {
		changes = append(changes, "required_observation_days")
	}
	return changes
}

func (c *TreatmentCase) ReviewPlan(reviewer, decision, comment string, rev int) error {
	version := 0
	if c.Plan != nil {
		version = c.Plan.Version
	}
	return c.ReviewPlanVersion(reviewer, decision, comment, version, rev)
}
func (c *TreatmentCase) ReviewPlanVersion(reviewer, decision, comment string, version, rev int) error {
	items := make([]PlanReviewItem, 0, 4)
	result := "PASS"
	if decision == "REJECT" {
		result = "FAIL"
	}
	for _, name := range reviewItemOrder {
		items = append(items, PlanReviewItem{Item: name, Result: result, Comment: comment})
	}
	return c.ReviewPlanChecklist(reviewer, decision, comment, items, version, rev)
}
func (c *TreatmentCase) ReviewPlanChecklist(reviewer, decision, comment string, items []PlanReviewItem, version, rev int) error {
	if err := c.check(rev); err != nil {
		return err
	}
	if c.Status != StatusAssessed || c.Plan == nil {
		return fmt.Errorf("%w: review", ErrForbidden)
	}
	if version != c.Plan.Version {
		return fmt.Errorf("%w: plan version", ErrConflict)
	}
	if reviewer == "" || reviewer == c.Plan.AuthorID {
		return fmt.Errorf("%w: reviewer separation", ErrInvalid)
	}
	if decision != "APPROVE" && decision != "REJECT" {
		return fmt.Errorf("%w: decision", ErrInvalid)
	}
	if decision == "REJECT" && strings.TrimSpace(comment) == "" {
		return fmt.Errorf("%w: review comment required", ErrInvalid)
	}
	ordered, err := normalizeReviewItems(items, decision)
	if err != nil {
		return err
	}
	c.Plan.ReviewerID = reviewer
	c.Plan.ReviewDecision = decision
	c.Plan.ReviewComment = comment
	c.Plan.ReviewItems = ordered
	c.bump()
	if decision == "APPROVE" {
		c.Status = StatusPlanApproved
	} else {
		c.Status = StatusPlanRevision
	}
	return nil
}

func (c *TreatmentCase) SubmitPilot(p PilotTrial, rev int) error {
	if err := c.check(rev); err != nil {
		return err
	}
	if c.Status != StatusPlanApproved {
		return fmt.Errorf("%w: pilot", ErrForbidden)
	}
	if p.ObservationDays <= 0 || p.ColorThreshold <= 0 || !finite(p.BeforeActivity) || !finite(p.AfterActivity) || !finite(p.BeforeColorDelta) || !finite(p.AfterColorDelta) || p.BeforeActivity < 0 || p.AfterActivity < 0 || p.BeforeColorDelta < 0 || p.AfterColorDelta < 0 {
		return fmt.Errorf("%w: pilot parameters", ErrInvalid)
	}
	providedObservations := len(p.Observations) > 0
	if !providedObservations {
		p.Observations = []PilotObservation{{ObservationDay: 1, Activity: p.BeforeActivity, ColorDelta: p.BeforeColorDelta}}
		if p.ObservationDays > 1 {
			p.Observations = append(p.Observations, PilotObservation{ObservationDay: p.ObservationDays, Activity: p.AfterActivity, ColorDelta: p.AfterColorDelta})
		}
	} else {
		for i, o := range p.Observations {
			if o.ObservationDay <= 0 || o.ObservationDay > p.ObservationDays || !finite(o.Activity) || !finite(o.ColorDelta) || o.Activity < 0 || o.ColorDelta < 0 {
				return fmt.Errorf("%w: invalid pilot observation", ErrInvalid)
			}
			if i > 0 && o.ObservationDay <= p.Observations[i-1].ObservationDay {
				return fmt.Errorf("%w: observations must be strictly ordered", ErrInvalid)
			}
		}
		if p.Observations[0].ObservationDay != 1 || p.Observations[len(p.Observations)-1].ObservationDay != p.ObservationDays {
			return fmt.Errorf("%w: first and last observation required", ErrInvalid)
		}
		p.BeforeActivity, p.BeforeColorDelta = p.Observations[0].Activity, p.Observations[0].ColorDelta
		p.AfterActivity, p.AfterColorDelta = p.Observations[len(p.Observations)-1].Activity, p.Observations[len(p.Observations)-1].ColorDelta
	}
	p.FailureReasons = nil
	p.Round = 1
	if c.Pilot != nil {
		p.Round = c.Pilot.Round + 1
	}
	p.RecordedAt = time.Now()
	p.ActivityGatePassed = p.BeforeActivity == 0 && p.AfterActivity == 0 || (p.BeforeActivity > 0 && (p.BeforeActivity-p.AfterActivity)/p.BeforeActivity >= 0.5)
	p.ColorGatePassed = true
	for _, o := range p.Observations {
		if o.ColorDelta > p.ColorThreshold {
			p.ColorGatePassed = false
			p.FailureReasons = append(p.FailureReasons, fmt.Sprintf("第%d日色差%.4g超过阈值%.4g", o.ObservationDay, o.ColorDelta, p.ColorThreshold))
		}
	}
	p.ObservationGatePassed = p.ObservationDays >= requiredObservation(c.Plan) && (!providedObservations || len(p.Observations) == p.ObservationDays)
	p.Passed = p.ActivityGatePassed && p.ColorGatePassed && p.ObservationGatePassed
	if !p.ActivityGatePassed {
		p.FailureReasons = append(p.FailureReasons, "活性下降比例未达标")
	}
	if !p.ObservationGatePassed {
		p.FailureReasons = append(p.FailureReasons, "观察周期不足")
	}
	if c.Plan != nil {
		p.PlanVersion = c.Plan.Version
	}
	if c.Pilot != nil {
		c.PilotHistory = append(c.PilotHistory, *c.Pilot)
	}
	c.Pilot = &p
	c.bump()
	if p.Passed {
		c.Status = StatusPilotPassed
	} else {
		c.Status = StatusPilotRevision
	}
	return nil
}
func requiredObservation(p *TreatmentPlan) int {
	if p != nil && p.RequiredObservationDays > 0 {
		return p.RequiredObservationDays
	}
	return 1
}

func (c *TreatmentCase) StartExecution(by string, rev int) error {
	if err := c.check(rev); err != nil {
		return err
	}
	if c.Status != StatusPilotPassed {
		return fmt.Errorf("%w: start", ErrForbidden)
	}
	if strings.TrimSpace(by) == "" {
		return ErrInvalid
	}
	c.Status, c.ExecutionResponsible = StatusInProgress, by
	c.bump()
	return nil
}
func (c *TreatmentCase) StartExecutionWithPlan(by string, plan []PlannedCheckpoint, rev int) error {
	if err := c.check(rev); err != nil {
		return err
	}
	if c.Status != StatusPilotPassed {
		return fmt.Errorf("%w: start", ErrForbidden)
	}
	if by == "" || len(plan) == 0 {
		return ErrInvalid
	}
	seen := map[string]bool{}
	for i, item := range plan {
		item.CheckpointID, item.Phase, item.ExpectedCondition, item.Unit = strings.TrimSpace(item.CheckpointID), strings.ToUpper(strings.TrimSpace(item.Phase)), strings.TrimSpace(item.ExpectedCondition), strings.TrimSpace(item.Unit)
		if item.CheckpointID == "" || seen[item.CheckpointID] || item.Sequence != i+1 || !validCheckpointPhase(item.Phase) || item.ExpectedCondition == "" || !validUnit(item.Unit) {
			return fmt.Errorf("%w: invalid planned_checkpoints", ErrInvalid)
		}
		seen[item.CheckpointID] = true
		plan[i] = item
	}
	c.PlannedCheckpoints = append([]PlannedCheckpoint(nil), plan...)
	c.Status = StatusInProgress
	c.ExecutionResponsible = by
	c.bump()
	return nil
}
func (c *TreatmentCase) RecordCheckpoint(cp ExecutionCheckpoint, rev int) error {
	return c.RecordCheckpoints([]ExecutionCheckpoint{cp}, rev)
}
func (c *TreatmentCase) RecordCheckpoints(checkpoints []ExecutionCheckpoint, rev int) error {
	if err := c.check(rev); err != nil {
		return err
	}
	if c.Status != StatusInProgress {
		return fmt.Errorf("%w: checkpoint", ErrForbidden)
	}
	if len(checkpoints) == 0 {
		return fmt.Errorf("%w: empty checkpoint batch", ErrInvalid)
	}
	stagedCheckpoints := append([]ExecutionCheckpoint(nil), c.Checkpoints...)
	stagedDeviations := append([]Deviation(nil), c.Deviations...)
	stagedRectifications := append([]RectificationItem(nil), c.RectificationItems...)
	sawDeviation := false
	for _, cp := range checkpoints {
		if cp.Sequence <= 0 || strings.TrimSpace(cp.ExpectedCondition) == "" || strings.TrimSpace(cp.ObservedValue) == "" || strings.TrimSpace(cp.RecordedBy) == "" || (cp.Result != "PASS" && cp.Result != "DEVIATION") {
			return ErrInvalid
		}
		if cp.Sequence != len(stagedCheckpoints)+1 {
			return fmt.Errorf("%w: checkpoint sequence", ErrInvalid)
		}
		if !validCheckpointPhase(cp.Phase) {
			return fmt.Errorf("%w: checkpoint phase", ErrInvalid)
		}
		if !validUnit(cp.Unit) {
			return fmt.Errorf("%w: measurement unit", ErrInvalid)
		}
		if cp.Unit != "" {
			v, err := strconv.ParseFloat(strings.TrimSpace(cp.ObservedValue), 64)
			if err != nil || !finite(v) {
				return fmt.Errorf("%w: observed value", ErrInvalid)
			}
		}
		condition := strings.ToLower(cp.ExpectedCondition)
		if strings.Contains(condition, "temperature") || strings.Contains(condition, "humidity") || strings.Contains(condition, "numeric") || strings.Contains(condition, "温度") || strings.Contains(condition, "湿度") {
			if cp.Unit == "" {
				return fmt.Errorf("%w: measurement unit required", ErrInvalid)
			}
			v, err := strconv.ParseFloat(strings.TrimSpace(cp.ObservedValue), 64)
			if err != nil || !finite(v) {
				return fmt.Errorf("%w: observed numeric value", ErrInvalid)
			}
		}
		if cp.RecordedAt.IsZero() {
			cp.RecordedAt = time.Now()
		}
		cp.CaseID = c.CaseID
		if cp.CheckpointID == "" {
			cp.CheckpointID = fmt.Sprintf("%s-cp-%03d", c.CaseID, cp.Sequence)
		}
		for _, existing := range stagedCheckpoints {
			if existing.CheckpointID == cp.CheckpointID {
				return fmt.Errorf("%w: checkpoint already reconciled", ErrConflict)
			}
		}
		if cp.RectificationItemID == "" {
			if len(c.PlannedCheckpoints) > 0 {
				if cp.Sequence > len(c.PlannedCheckpoints) {
					return fmt.Errorf("%w: checkpoint not in locked plan", ErrConflict)
				}
				planned := c.PlannedCheckpoints[cp.Sequence-1]
				if cp.CheckpointID != planned.CheckpointID || cp.Phase != planned.Phase || cp.ExpectedCondition != planned.ExpectedCondition || cp.Unit != planned.Unit {
					return fmt.Errorf("%w: checkpoint does not match locked plan", ErrConflict)
				}
			}
		} else {
			idx, ok := -1, false
			for i := range stagedRectifications {
				if stagedRectifications[i].ID == cp.RectificationItemID && stagedRectifications[i].Status == "OPEN" {
					idx, ok = i, true
					break
				}
			}
			if !ok {
				return fmt.Errorf("%w: rectification item unavailable", ErrInvalid)
			}
			for _, existing := range stagedCheckpoints {
				if existing.RectificationItemID == cp.RectificationItemID {
					return fmt.Errorf("%w: rectification item already reconciled", ErrConflict)
				}
			}
			if cp.Result == "PASS" {
				stagedRectifications[idx].CheckpointID = cp.CheckpointID
			}
		}
		if cp.Result == "DEVIATION" {
			if cp.DeviationID == "" {
				return fmt.Errorf("%w: deviation id", ErrInvalid)
			}
			for _, d := range stagedDeviations {
				if d.ID == cp.DeviationID {
					return fmt.Errorf("%w: duplicate deviation", ErrInvalid)
				}
			}
			rootID, recurrence := cp.DeviationID, 0
			if cp.RelatedDeviationID != "" {
				var related *Deviation
				for i := range stagedDeviations {
					if stagedDeviations[i].ID == cp.RelatedDeviationID {
						related = &stagedDeviations[i]
						break
					}
				}
				if related == nil || !related.Resolved || related.Type != cp.DeviationType {
					return fmt.Errorf("%w: invalid related_deviation_id", ErrInvalid)
				}
				rootID, recurrence = related.RootDeviationID, related.RecurrenceCount+1
				if rootID == "" {
					rootID = related.ID
				}
				if rootID == cp.DeviationID {
					return fmt.Errorf("%w: deviation recurrence cycle", ErrInvalid)
				}
			}
			sawDeviation = true
			created := cp.RecordedAt
			severity := deviationSeverity(cp.DeviationType, cp.ObservedValue)
			if recurrence == 1 && severity == "LOW" {
				severity = "MEDIUM"
			}
			if recurrence >= 2 {
				severity = "HIGH"
			}
			basis := ""
			if recurrence > 0 {
				basis = fmt.Sprintf("同根偏差第%d次复发，按复发次数升级", recurrence)
			}
			stagedDeviations = append(stagedDeviations, Deviation{ID: cp.DeviationID, Type: cp.DeviationType, Description: cp.ObservedValue, Severity: severity, ReportedBy: cp.RecordedBy, CreatedAt: &created, RelatedDeviationID: cp.RelatedDeviationID, RootDeviationID: rootID, RecurrenceCount: recurrence, EscalationBasis: basis})
		} else if sawDeviation {
			return fmt.Errorf("%w: checkpoint after open deviation", ErrInvalid)
		}
		stagedCheckpoints = append(stagedCheckpoints, cp)
	}
	c.Checkpoints, c.Deviations, c.RectificationItems = stagedCheckpoints, stagedDeviations, stagedRectifications
	if sawDeviation {
		c.Status = StatusPaused
	}
	c.bump()
	return nil
}
func deviationSeverity(kind, observed string) string {
	k := strings.ToUpper(strings.TrimSpace(kind))
	if strings.Contains(k, "SAFETY") || strings.Contains(k, "MATERIAL") || strings.Contains(k, "HIGH") || strings.Contains(kind, "安全") || strings.Contains(kind, "材料") {
		return "HIGH"
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(observed), 64)
	if err == nil {
		if math.Abs(v) >= 10 {
			return "HIGH"
		}
		if math.Abs(v) >= 5 {
			return "MEDIUM"
		}
	}
	if strings.Contains(k, "MEDIUM") {
		return "MEDIUM"
	}
	return "LOW"
}
func (c *TreatmentCase) ResolveDeviation(id, evidence, by string, rev int) error {
	refs := []string{}
	if strings.TrimSpace(evidence) != "" {
		refs = append(refs, evidence)
	}
	return c.ResolveDeviationWithEvidence(id, evidence, refs, "", by, rev)
}
func (c *TreatmentCase) ResolveDeviationWithEvidence(id, action string, refs []string, reviewer, by string, rev int) error {
	if err := c.check(rev); err != nil {
		return err
	}
	for i := range c.Deviations {
		if c.Deviations[i].ID == id {
			if c.Deviations[i].Resolved {
				return fmt.Errorf("%w: deviation already resolved", ErrInvalid)
			}
			if c.Status != StatusPaused {
				return fmt.Errorf("%w: resolve", ErrForbidden)
			}
			if strings.TrimSpace(action) == "" || by == "" || by == c.Deviations[i].ReportedBy {
				return ErrInvalid
			}
			clean := make([]string, 0, len(refs))
			for _, ref := range refs {
				if strings.TrimSpace(ref) == "" {
					return fmt.Errorf("%w: evidence reference", ErrInvalid)
				}
				clean = append(clean, strings.TrimSpace(ref))
			}
			if len(clean) == 0 {
				return fmt.Errorf("%w: evidence required", ErrInvalid)
			}
			if c.Deviations[i].Severity == "HIGH" && (len(clean) < 2 || strings.TrimSpace(reviewer) == "" || reviewer == by || reviewer == c.Deviations[i].ReportedBy) {
				return fmt.Errorf("%w: high deviation requires independent reviewer and two evidence references", ErrInvalid)
			}
			now := time.Now()
			c.Deviations[i].CorrectiveEvidence = strings.Join(clean, ",")
			c.Deviations[i].CorrectiveAction = action
			c.Deviations[i].EvidenceRefs = clean
			c.Deviations[i].ReviewerID = reviewer
			c.Deviations[i].ResolvedBy = by
			c.Deviations[i].Resolved = true
			c.Deviations[i].ResolvedAt = &now
			all := true
			for _, d := range c.Deviations {
				if !d.Resolved {
					all = false
					break
				}
			}
			if all {
				c.Status = StatusInProgress
			}
			for j := range c.Checkpoints {
				if c.Checkpoints[j].DeviationID == id {
					c.Checkpoints[j].Result = "PASS"
					c.Checkpoints[j].EvidenceRefs = clean
					if c.Checkpoints[j].RectificationItemID != "" {
						for k := range c.RectificationItems {
							if c.RectificationItems[k].ID == c.Checkpoints[j].RectificationItemID {
								c.RectificationItems[k].CheckpointID = c.Checkpoints[j].CheckpointID
							}
						}
					}
				}
			}
			c.bump()
			return nil
		}
	}
	return fmt.Errorf("%w: deviation", ErrInvalid)
}
func (c *TreatmentCase) CompleteExecution(rev int) error {
	if err := c.check(rev); err != nil {
		return err
	}
	if c.Status != StatusInProgress {
		return fmt.Errorf("%w: complete", ErrForbidden)
	}
	if len(c.Checkpoints) == 0 {
		return fmt.Errorf("%w: no checkpoints", ErrInvalid)
	}
	if len(c.PlannedCheckpoints) > 0 {
		if len(c.Checkpoints) < len(c.PlannedCheckpoints) {
			return fmt.Errorf("%w: missing planned checkpoint %s", ErrInvalid, c.PlannedCheckpoints[len(c.Checkpoints)].CheckpointID)
		}
		for i, planned := range c.PlannedCheckpoints {
			cp := c.Checkpoints[i]
			if cp.CheckpointID != planned.CheckpointID || cp.Result != "PASS" {
				return fmt.Errorf("%w: planned checkpoint %s pending", ErrInvalid, planned.CheckpointID)
			}
		}
	}
	for _, item := range c.RectificationItems {
		if item.Status == "OPEN" && item.CheckpointID == "" {
			return fmt.Errorf("%w: rectification item %s pending", ErrInvalid, item.ID)
		}
	}
	for _, d := range c.Deviations {
		if !d.Resolved {
			return fmt.Errorf("%w: open deviation", ErrInvalid)
		}
	}
	for i, cp := range c.Checkpoints {
		if cp.Sequence != i+1 || cp.Result != "PASS" {
			return fmt.Errorf("%w: checkpoint %d pending", ErrInvalid, i+1)
		}
	}
	c.CheckpointSummary = c.CheckpointSummary[:0]
	c.CheckpointCompletionSummary = c.CheckpointCompletionSummary[:0]
	for _, cp := range c.Checkpoints {
		c.CheckpointSummary = append(c.CheckpointSummary, fmt.Sprintf("%03d:%s:%s:%s", cp.Sequence, cp.CheckpointID, cp.Phase, cp.Result))
		c.CheckpointCompletionSummary = append(c.CheckpointCompletionSummary, CheckpointCompletion{CheckpointID: cp.CheckpointID, Sequence: cp.Sequence, Phase: cp.Phase, ExpectedCondition: cp.ExpectedCondition, ObservedValue: cp.ObservedValue, Unit: cp.Unit, Result: cp.Result})
	}
	c.Status = StatusTreatmentCompleted
	c.bump()
	return nil
}
func (c *TreatmentCase) VerifyOutcome(o OutcomeVerification, rev int) error {
	if err := c.check(rev); err != nil {
		return err
	}
	if c.Status != StatusTreatmentCompleted {
		return fmt.Errorf("%w: outcome", ErrForbidden)
	}
	if o.VerifiedBy == "" || o.ObservationDays <= 0 || o.MeasuredAt.IsZero() || !finite(o.PostActivity) || !finite(o.ColorDelta) || !finite(o.SurfaceStability) || o.ActivityThreshold <= 0 || o.ColorThreshold <= 0 || o.StabilityThreshold <= 0 {
		return ErrInvalid
	}
	activityPassed := o.PostActivity <= o.ActivityThreshold
	colorPassed := o.ColorDelta <= o.ColorThreshold
	stabilityPassed := o.SurfaceStability >= o.StabilityThreshold
	observationPassed := o.ObservationDays >= requiredObservation(c.Plan)
	o.Passed = activityPassed && colorPassed && stabilityPassed && observationPassed
	o.Round = 1
	if c.Outcome != nil {
		o.Round = c.Outcome.Round + 1
	}
	o.ThresholdSnapshot = map[string]float64{"activity": o.ActivityThreshold, "color": o.ColorThreshold, "stability": o.StabilityThreshold, "observation_days": float64(requiredObservation(c.Plan))}
	if !o.Passed {
		if o.PostActivity > o.ActivityThreshold {
			o.FailureItems = append(o.FailureItems, "活性")
		}
		if o.ColorDelta > o.ColorThreshold {
			o.FailureItems = append(o.FailureItems, "色差")
		}
		if o.SurfaceStability < o.StabilityThreshold {
			o.FailureItems = append(o.FailureItems, "表面稳定性")
		}
		if !observationPassed {
			o.FailureItems = append(o.FailureItems, "观察周期")
		}
		if strings.TrimSpace(o.Rectification) == "" {
			return fmt.Errorf("%w: rectification required", ErrInvalid)
		}
	}
	o.VerifiedAt = time.Now()
	if c.Outcome != nil {
		c.OutcomeHistory = append(c.OutcomeHistory, *c.Outcome)
	}
	c.Outcome = &o
	if o.Passed {
		for i := range c.RectificationItems {
			if c.RectificationItems[i].Status == "OPEN" {
				c.RectificationItems[i].Status = "CLOSED"
				c.RectificationItems[i].ClosedByOutcomeRound = o.Round
			}
		}
	} else {
		value := map[string]float64{"活性": o.PostActivity, "色差": o.ColorDelta, "表面稳定性": o.SurfaceStability, "观察周期": float64(o.ObservationDays)}
		target := map[string]float64{"活性": o.ActivityThreshold, "色差": o.ColorThreshold, "表面稳定性": o.StabilityThreshold, "观察周期": float64(requiredObservation(c.Plan))}
		for i, metric := range o.FailureItems {
			c.RectificationItems = append(c.RectificationItems, RectificationItem{ID: fmt.Sprintf("rect-%d-%02d", o.Round, i+1), OutcomeRound: o.Round, Metric: metric, FailureValue: value[metric], TargetThreshold: target[metric], Requirement: strings.TrimSpace(o.Rectification), Status: "OPEN"})
		}
	}
	c.bump()
	if o.Passed {
		c.Status = StatusOutcomeVerified
	} else {
		c.Status = StatusInProgress
	}
	return nil
}
func (c *TreatmentCase) Archive(manifest EvidenceManifest, by string, rev int) error {
	if err := c.check(rev); err != nil {
		return err
	}
	missing := []string{}
	if c.Assessment == nil {
		missing = append(missing, "assessment")
	}
	if c.Plan == nil || c.Plan.ReviewDecision != "APPROVE" {
		missing = append(missing, "approved_plan")
	}
	if c.Pilot == nil || !c.Pilot.Passed {
		missing = append(missing, "pilot")
	}
	if len(c.Checkpoints) == 0 {
		missing = append(missing, "checkpoints")
	}
	if c.Outcome == nil || !c.Outcome.Passed {
		missing = append(missing, "outcome")
	}
	if c.Status != StatusOutcomeVerified || len(missing) > 0 {
		return fmt.Errorf("%w: archive missing %s", ErrForbidden, strings.Join(missing, ","))
	}
	for _, d := range c.Deviations {
		if !d.Resolved || strings.TrimSpace(d.CorrectiveEvidence) == "" {
			return fmt.Errorf("%w: missing corrective evidence", ErrInvalid)
		}
	}
	if by == "" || len(manifest.EvidenceItems) == 0 || manifest.EventChainHead == "" || manifest.ContentDigest == "" {
		return ErrInvalid
	}
	manifest.CaseID = c.CaseID
	manifest.CaseRevision = c.Revision
	manifest.SealedBy = by
	manifest.SealedAt = time.Now()
	c.Evidence = &manifest
	c.Status = StatusArchived
	c.ArchivedAt = &manifest.SealedAt
	c.bump()
	return nil
}
