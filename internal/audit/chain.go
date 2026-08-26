package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mural-biocare/internal/domain"
	"sort"
	"time"
)

type Event struct {
	CaseID         string    `json:"case_id"`
	Revision       int       `json:"revision"`
	Type           string    `json:"type"`
	Actor          string    `json:"actor"`
	Payload        any       `json:"payload"`
	At             time.Time `json:"at"`
	PreviousDigest string    `json:"previous_digest"`
	Digest         string    `json:"digest"`
}

func DigestEvent(e Event) string {
	e.Digest = ""
	b, _ := json.Marshal(e)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func Append(events []Event, caseID string, rev int, typ, actor string, payload any, at time.Time) Event {
	prev := ""
	if len(events) > 0 {
		prev = events[len(events)-1].Digest
	}
	e := Event{CaseID: caseID, Revision: rev, Type: typ, Actor: actor, Payload: payload, At: at, PreviousDigest: prev}
	e.Digest = DigestEvent(e)
	return e
}
func Verify(events []Event) error {
	prev := ""
	for i, e := range events {
		if e.PreviousDigest != prev {
			return fmt.Errorf("event chain break at %d", i)
		}
		if DigestEvent(e) != e.Digest {
			return fmt.Errorf("event digest mismatch at %d", i)
		}
		if i > 0 && e.Revision != events[i-1].Revision+1 {
			return fmt.Errorf("revision jump at %d", i)
		}
		prev = e.Digest
	}
	return nil
}
func Manifest(c *domain.TreatmentCase, events []Event, by string) (domain.EvidenceManifest, error) {
	if err := Verify(events); err != nil {
		return domain.EvidenceManifest{}, err
	}
	if len(events) == 0 {
		return domain.EvidenceManifest{}, fmt.Errorf("missing events")
	}
	items := evidenceItems(events)
	cc := *c
	cc.Evidence = nil
	raw, _ := json.Marshal(struct {
		Case  *domain.TreatmentCase
		Items []string
	}{&cc, items})
	h := sha256.Sum256(raw)
	return domain.EvidenceManifest{ManifestID: fmt.Sprintf("manifest-%d", time.Now().UnixNano()), CaseID: c.CaseID, CaseRevision: c.Revision, EvidenceItems: items, EvidenceIndex: evidenceIndex(events), EventChainHead: events[len(events)-1].Digest, ContentDigest: hex.EncodeToString(h[:]), SealedBy: by, SealedAt: time.Now()}, nil
}
func evidenceItems(events []Event) []string {
	items := make([]string, 0, len(events))
	for _, e := range events {
		items = append(items, fmt.Sprintf("%06d:%s:%s", e.Revision, e.Type, e.Digest))
	}
	sort.Strings(items)
	return items
}
func evidenceIndex(events []Event) map[string][]string {
	index := map[string][]string{"profile": {}, "assessment": {}, "plan": {}, "pilot": {}, "execution": {}, "outcome": {}}
	seen := map[string]map[string]bool{}
	for key := range index {
		seen[key] = map[string]bool{}
	}
	for _, e := range events {
		category := ""
		switch e.Type {
		case "CASE_CREATED", "PROFILE_CORRECTED", "BASELINE_RETESTED":
			category = "profile"
		case "ASSESSMENT_SUBMITTED":
			category = "assessment"
		case "PLAN_SUBMITTED", "PLAN_REVIEWED":
			category = "plan"
		case "PILOT_RECORDED":
			category = "pilot"
		case "EXECUTION_STARTED", "CHECKPOINT_BATCH_RECORDED", "DEVIATION_RESOLVED", "EXECUTION_COMPLETED":
			category = "execution"
		case "OUTCOME_VERIFIED":
			category = "outcome"
		}
		if category == "" {
			continue
		}
		item := fmt.Sprintf("%06d:%s:%s", e.Revision, e.Type, e.Digest)
		if !seen[category][item] {
			index[category] = append(index[category], item)
			seen[category][item] = true
		}
	}
	for key := range index {
		sort.Strings(index[key])
	}
	return index
}
func VerifyManifest(c *domain.TreatmentCase, events []Event) error {
	if c.Evidence == nil {
		return fmt.Errorf("missing manifest")
	}
	if c.Status != domain.StatusArchived || c.ArchivedAt == nil || !c.ArchivedAt.Equal(c.Evidence.SealedAt) {
		return fmt.Errorf("manifest sealed state mismatch")
	}
	if err := Verify(events); err != nil {
		return err
	}
	head := -1
	for i, e := range events {
		if e.Digest == c.Evidence.EventChainHead {
			head = i
			break
		}
	}
	if head < 0 || c.Evidence.ContentDigest == "" {
		return fmt.Errorf("manifest chain head mismatch")
	}
	if head+1 >= len(events) || events[head+1].Type != "CASE_ARCHIVED" || head+1 != len(events)-1 {
		return fmt.Errorf("manifest archive event missing")
	}
	var sealed domain.EvidenceManifest
	payload, payloadErr := json.Marshal(events[head+1].Payload)
	if payloadErr != nil || json.Unmarshal(payload, &sealed) != nil {
		return fmt.Errorf("manifest archive payload mismatch")
	}
	want, _ := json.Marshal(c.Evidence)
	got, _ := json.Marshal(&sealed)
	if string(want) != string(got) {
		return fmt.Errorf("manifest sealed metadata mismatch")
	}
	items := evidenceItems(events[:head+1])
	if len(items) != len(c.Evidence.EvidenceItems) {
		return fmt.Errorf("manifest evidence item missing")
	}
	for i := range items {
		if items[i] != c.Evidence.EvidenceItems[i] {
			return fmt.Errorf("manifest evidence item mismatch at %d", i)
		}
	}
	wantIndex, gotIndex := evidenceIndex(events[:head+1]), c.Evidence.EvidenceIndex
	wantIndexJSON, _ := json.Marshal(wantIndex)
	gotIndexJSON, _ := json.Marshal(gotIndex)
	if string(wantIndexJSON) != string(gotIndexJSON) {
		return fmt.Errorf("manifest evidence index mismatch")
	}
	cc := *c
	cc.Evidence = nil
	cc.Status = domain.StatusOutcomeVerified
	cc.Revision = c.Evidence.CaseRevision
	cc.ArchivedAt = nil
	raw, _ := json.Marshal(struct {
		Case  *domain.TreatmentCase
		Items []string
	}{&cc, items})
	h := sha256.Sum256(raw)
	if hex.EncodeToString(h[:]) != c.Evidence.ContentDigest {
		return fmt.Errorf("manifest digest mismatch")
	}
	return nil
}
