package application

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mural-biocare/internal/audit"
	"mural-biocare/internal/domain"
	"mural-biocare/internal/persistence"
	"time"
)

type Service struct {
	Store                  *persistence.Store
	Locker                 *persistence.CaseLocker
	BaselineRetestInterval time.Duration
}

func New(s *persistence.Store) *Service {
	return &Service{Store: s, Locker: persistence.NewCaseLocker(), BaselineRetestInterval: 24 * time.Hour}
}

type Command struct {
	RequestID, Actor string
	ExpectedRevision int
	Payload          any
}

func (s *Service) mutate(id string, cmd Command, typ string, fn func(*domain.TreatmentCase) error) (*domain.TreatmentCase, error) {
	unlock := s.Locker.Lock(id)
	defer unlock()
	if cmd.RequestID != "" {
		if old, ok := s.Store.GetIdempotency(cmd.RequestID); ok {
			if old.Fingerprint != commandFingerprint(id, typ, cmd) {
				return nil, fmt.Errorf("request_id conflict")
			}
			var c domain.TreatmentCase
			json.Unmarshal(old.Response, &c)
			return &c, nil
		}
	}
	c, ok := s.Store.Get(id)
	if !ok {
		return nil, fmt.Errorf("case not found")
	}
	if err := fn(c); err != nil {
		return nil, err
	}
	payload := cmd.Payload
	switch typ {
	case "PROFILE_CORRECTED":
		payload = c.ProfileCorrections[len(c.ProfileCorrections)-1]
	case "BASELINE_RETESTED":
		payload = map[string]any{"measurement": c.BaselineMeasurements[len(c.BaselineMeasurements)-1], "baseline_status": c.BaselineStatus, "unlocked_at": c.BaselineUnlockedAt, "unlock_basis": "连续两次 NORMAL 且间隔达到配置周期"}
	case "ASSESSMENT_SUBMITTED":
		payload = map[string]any{"method": c.Assessment.Method, "sample_count": len(c.Assessment.SamplePoints), "conclusion": c.Assessment.OrganismFindings, "version": c.Assessment.Version, "diff": latestAssessmentDiff(c)}
	case "PLAN_SUBMITTED":
		payload = c.Plan
	case "PLAN_REVIEWED":
		payload = map[string]any{"plan_version": c.Plan.Version, "reviewer_id": c.Plan.ReviewerID, "review_items": c.Plan.ReviewItems, "decision": c.Plan.ReviewDecision, "comment": c.Plan.ReviewComment}
	case "PILOT_RECORDED":
		payload = c.Pilot
	case "EXECUTION_STARTED":
		payload = map[string]any{"execution_responsible": c.ExecutionResponsible, "planned_checkpoints": c.PlannedCheckpoints}
	case "CHECKPOINT_BATCH_RECORDED":
		payload = map[string]any{"checkpoints": c.Checkpoints, "deviations": c.Deviations, "rectification_items": c.RectificationItems}
	case "DEVIATION_RESOLVED":
		for i := len(c.Deviations) - 1; i >= 0; i-- {
			if c.Deviations[i].Resolved {
				payload = c.Deviations[i]
				break
			}
		}
	case "OUTCOME_VERIFIED":
		payload = c.Outcome
	case "CASE_ARCHIVED":
		payload = c.Evidence
	}
	ev := audit.Append(s.Store.Events(id), id, c.Revision, typ, cmd.Actor, payload, time.Now())
	if cmd.RequestID != "" {
		b, _ := json.Marshal(c)
		if err := s.Store.SaveIdempotency(cmd.RequestID, commandFingerprint(id, typ, cmd), b); err != nil {
			return nil, err
		}
	}
	if err := s.Store.Save(c, ev); err != nil {
		if cmd.RequestID != "" {
			s.Store.RemoveIdempotency(cmd.RequestID)
		}
		return nil, err
	}
	return c, nil
}
func latestAssessmentDiff(c *domain.TreatmentCase) any {
	if len(c.AssessmentDiffs) == 0 {
		return nil
	}
	return c.AssessmentDiffs[len(c.AssessmentDiffs)-1]
}
func commandFingerprint(id, typ string, cmd Command) string {
	return fingerprint(struct {
		CaseID, Type, Actor string
		ExpectedRevision    int
		Payload             any
	}{id, typ, cmd.Actor, cmd.ExpectedRevision, cmd.Payload})
}
func fingerprint(v any) string {
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func (s *Service) Create(site, section, symptom, owner string, temp, humidity float64, reqID string) (*domain.TreatmentCase, error) {
	createPayload := struct {
		Site, Section, Symptom, Owner string
		Temperature, Humidity         float64
	}{site, section, symptom, owner, temp, humidity}
	if reqID != "" {
		if old, ok := s.Store.GetIdempotency(reqID); ok {
			if old.Fingerprint != fingerprint(createPayload) {
				return nil, fmt.Errorf("request_id conflict")
			}
			var existing domain.TreatmentCase
			if json.Unmarshal(old.Response, &existing) == nil {
				return &existing, nil
			}
		}
	}
	id := fmt.Sprintf("case-%d", time.Now().UnixNano())
	c, err := domain.NewCase(id, site, section, symptom, owner, temp, humidity, time.Now())
	if err != nil {
		return nil, err
	}
	ev := audit.Append(nil, id, c.Revision, "CASE_CREATED", owner, c, time.Now())
	if reqID != "" {
		b, _ := json.Marshal(c)
		if err = s.Store.SaveIdempotency(reqID, fingerprint(createPayload), b); err != nil {
			return nil, err
		}
	}
	if err = s.Store.Save(c, ev); err != nil {
		if reqID != "" {
			s.Store.RemoveIdempotency(reqID)
		}
		return nil, err
	}
	return c, nil
}
func (s *Service) CorrectProfile(id string, in domain.ProfileCorrectionInput, cmd Command) (*domain.TreatmentCase, error) {
	return s.mutate(id, cmd, "PROFILE_CORRECTED", func(c *domain.TreatmentCase) error {
		if c.OwnerID != cmd.Actor {
			return ErrUnauthorized
		}
		return c.CorrectProfile(in, cmd.Actor, cmd.ExpectedRevision)
	})
}
func (s *Service) Assessment(id string, a domain.ContaminationAssessment, cmd Command) (*domain.TreatmentCase, error) {
	return s.mutate(id, cmd, "ASSESSMENT_SUBMITTED", func(c *domain.TreatmentCase) error {
		if a.AssessorID == "" {
			a.AssessorID = cmd.Actor
		}
		if cmd.Actor == "" || a.AssessorID != cmd.Actor {
			return fmt.Errorf("%w: assessor identity", domain.ErrInvalid)
		}
		return c.SubmitAssessment(a, cmd.ExpectedRevision)
	})
}
func (s *Service) UpdateBaseline(id string, temp, humidity float64, measuredAt time.Time, cmd Command) (*domain.TreatmentCase, error) {
	return s.mutate(id, cmd, "BASELINE_RETESTED", func(c *domain.TreatmentCase) error {
		return c.UpdateBaselineWithInterval(temp, humidity, measuredAt, cmd.Actor, cmd.ExpectedRevision, s.BaselineRetestInterval)
	})
}
func (s *Service) Plan(id string, p domain.TreatmentPlan, cmd Command) (*domain.TreatmentCase, error) {
	return s.mutate(id, cmd, "PLAN_SUBMITTED", func(c *domain.TreatmentCase) error {
		p.AuthorID = cmd.Actor
		return c.SubmitPlan(p, cmd.ExpectedRevision)
	})
}
func (s *Service) Review(id string, decision, comment string, cmd Command) (*domain.TreatmentCase, error) {
	return s.ReviewVersion(id, 0, decision, comment, cmd)
}
func (s *Service) ReviewVersion(id string, version int, decision, comment string, cmd Command) (*domain.TreatmentCase, error) {
	return s.ReviewChecklist(id, version, decision, comment, nil, cmd)
}
func (s *Service) ReviewChecklist(id string, version int, decision, comment string, items []domain.PlanReviewItem, cmd Command) (*domain.TreatmentCase, error) {
	return s.mutate(id, cmd, "PLAN_REVIEWED", func(c *domain.TreatmentCase) error {
		if version == 0 && c.Plan != nil {
			version = c.Plan.Version
		}
		if items == nil {
			return c.ReviewPlanVersion(cmd.Actor, decision, comment, version, cmd.ExpectedRevision)
		}
		return c.ReviewPlanChecklist(cmd.Actor, decision, comment, items, version, cmd.ExpectedRevision)
	})
}
func (s *Service) Pilot(id string, p domain.PilotTrial, cmd Command) (*domain.TreatmentCase, error) {
	return s.mutate(id, cmd, "PILOT_RECORDED", func(c *domain.TreatmentCase) error { return c.SubmitPilot(p, cmd.ExpectedRevision) })
}
func (s *Service) Start(id string, cmd Command) (*domain.TreatmentCase, error) {
	return s.StartWithPlan(id, nil, cmd)
}
func (s *Service) StartWithPlan(id string, plan []domain.PlannedCheckpoint, cmd Command) (*domain.TreatmentCase, error) {
	return s.mutate(id, cmd, "EXECUTION_STARTED", func(c *domain.TreatmentCase) error {
		if plan == nil {
			return c.StartExecution(cmd.Actor, cmd.ExpectedRevision)
		}
		return c.StartExecutionWithPlan(cmd.Actor, plan, cmd.ExpectedRevision)
	})
}
func (s *Service) Checkpoint(id string, cp domain.ExecutionCheckpoint, cmd Command) (*domain.TreatmentCase, error) {
	return s.Checkpoints(id, []domain.ExecutionCheckpoint{cp}, cmd)
}
func (s *Service) Checkpoints(id string, checkpoints []domain.ExecutionCheckpoint, cmd Command) (*domain.TreatmentCase, error) {
	return s.mutate(id, cmd, "CHECKPOINT_BATCH_RECORDED", func(c *domain.TreatmentCase) error { return c.RecordCheckpoints(checkpoints, cmd.ExpectedRevision) })
}
func (s *Service) Resolve(id, deviation, evidence string, cmd Command) (*domain.TreatmentCase, error) {
	return s.mutate(id, cmd, "DEVIATION_RESOLVED", func(c *domain.TreatmentCase) error {
		return c.ResolveDeviation(deviation, evidence, cmd.Actor, cmd.ExpectedRevision)
	})
}
func (s *Service) ResolveWithEvidence(id, deviation, action string, refs []string, reviewer string, cmd Command) (*domain.TreatmentCase, error) {
	return s.mutate(id, cmd, "DEVIATION_RESOLVED", func(c *domain.TreatmentCase) error {
		return c.ResolveDeviationWithEvidence(deviation, action, refs, reviewer, cmd.Actor, cmd.ExpectedRevision)
	})
}
func (s *Service) Complete(id string, cmd Command) (*domain.TreatmentCase, error) {
	return s.mutate(id, cmd, "EXECUTION_COMPLETED", func(c *domain.TreatmentCase) error { return c.CompleteExecution(cmd.ExpectedRevision) })
}
func (s *Service) Outcome(id string, o domain.OutcomeVerification, cmd Command) (*domain.TreatmentCase, error) {
	return s.mutate(id, cmd, "OUTCOME_VERIFIED", func(c *domain.TreatmentCase) error {
		if c.OwnerID == cmd.Actor || c.ExecutionResponsible == cmd.Actor {
			return fmt.Errorf("verifier separation")
		}
		o.VerifiedBy = cmd.Actor
		if o.MeasuredAt.IsZero() {
			o.MeasuredAt = time.Now()
		}
		return c.VerifyOutcome(o, cmd.ExpectedRevision)
	})
}
func (s *Service) Archive(id string, cmd Command) (*domain.TreatmentCase, error) {
	return s.mutate(id, cmd, "CASE_ARCHIVED", func(c *domain.TreatmentCase) error {
		events := s.Store.Events(id)
		readiness := archiveReadiness(c, events)
		if !readiness.Ready {
			return fmt.Errorf("%w: ARCHIVE_NOT_READY: %s", domain.ErrForbidden, readiness.Summary())
		}
		m, err := audit.Manifest(c, events, cmd.Actor)
		if err != nil {
			return err
		}
		return c.Archive(m, cmd.Actor, cmd.ExpectedRevision)
	})
}
func (s *Service) Get(id string) (*domain.TreatmentCase, bool) { return s.Store.Get(id) }
func (s *Service) Timeline(id string) []audit.Event            { return s.Store.Events(id) }
func (s *Service) Verify(id string) error {
	c, ok := s.Store.Get(id)
	if !ok {
		return fmt.Errorf("case not found")
	}
	return audit.VerifyManifest(c, s.Store.Events(id))
}

type VerificationReport struct {
	Valid      bool       `json:"valid"`
	SealedBy   string     `json:"sealed_by,omitempty"`
	SealedAt   *time.Time `json:"sealed_at,omitempty"`
	VerifiedAt time.Time  `json:"verified_at"`
	Reasons    []string   `json:"reasons,omitempty"`
}

func (s *Service) Verification(id string) (VerificationReport, error) {
	r := VerificationReport{VerifiedAt: time.Now()}
	c, ok := s.Store.Get(id)
	if !ok {
		r.Reasons = []string{"case not found"}
		return r, fmt.Errorf("case not found")
	}
	if c.Evidence != nil {
		r.SealedBy = c.Evidence.SealedBy
		r.SealedAt = &c.Evidence.SealedAt
	}
	if err := audit.VerifyManifest(c, s.Store.Events(id)); err != nil {
		r.Reasons = []string{err.Error()}
		return r, err
	}
	r.Valid = true
	return r, nil
}
