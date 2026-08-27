package cross_case_commit_owner_test

import (
	"encoding/json"
	"mural-biocare/internal/audit"
	"mural-biocare/internal/domain"
	"mural-biocare/internal/persistence"
	"sync"
	"testing"
	"time"
)

type gatedPayload struct {
	mu      sync.Mutex
	active  bool
	blocked bool
	entered chan struct{}
	release chan struct{}
}

func (p *gatedPayload) enable() {
	p.mu.Lock()
	p.active = true
	p.mu.Unlock()
}

func (p *gatedPayload) MarshalJSON() ([]byte, error) {
	p.mu.Lock()
	shouldBlock := p.active && !p.blocked
	if shouldBlock {
		p.blocked = true
	}
	p.mu.Unlock()
	if shouldBlock {
		close(p.entered)
		<-p.release
	}
	return json.Marshal("case-a-payload")
}

func TestConcurrentCaseCommitsKeepTransactionOwnership(t *testing.T) {
	store, err := persistence.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 26, 10, 0, 0, 0, time.UTC)
	caseA, err := domain.NewCase("case-a", "site-a", "section-a", "symptom-a", "owner-a", 20, 60, now)
	if err != nil {
		t.Fatal(err)
	}
	caseB, err := domain.NewCase("case-b", "site-b", "section-b", "symptom-b", "owner-b", 21, 61, now)
	if err != nil {
		t.Fatal(err)
	}
	payload := &gatedPayload{entered: make(chan struct{}), release: make(chan struct{})}
	eventA := audit.Append(nil, caseA.CaseID, caseA.Revision, "CASE_CREATED", caseA.OwnerID, payload, now)
	payload.enable()
	eventB := audit.Append(nil, caseB.CaseID, caseB.Revision, "CASE_CREATED", caseB.OwnerID, "case-b-payload", now)

	resultA := make(chan error, 1)
	go func() {
		resultA <- store.Save(caseA, eventA)
	}()
	<-payload.entered

	if err := store.Save(caseB, eventB); err != nil {
		t.Fatalf("case B save failed: %v", err)
	}
	close(payload.release)
	if err := <-resultA; err != nil {
		t.Fatalf("case A save returned an error instead of exposing the ownership bug: %v", err)
	}
	if _, ok := store.Get(caseA.CaseID); !ok {
		t.Fatalf("case A save reported success but the committed case is missing")
	}
}
