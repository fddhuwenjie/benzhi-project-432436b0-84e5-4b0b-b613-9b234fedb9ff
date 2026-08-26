package persistence

import (
	"bufio"
	"encoding/json"
	"fmt"
	"mural-biocare/internal/audit"
	"mural-biocare/internal/domain"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	dir     string
	mu      sync.Mutex
	cases   map[string]*domain.TreatmentCase
	events  map[string][]audit.Event
	idem    map[string]Idempotent
	pending *saveTransaction
}

type saveTransaction struct {
	caseID      string
	caseValue   *domain.TreatmentCase
	nextEvents  []audit.Event
	tmp         string
	final       string
	eventsTmp   string
	eventsFinal string
}
type Idempotent struct {
	Fingerprint string `json:"fingerprint"`
	Response    []byte `json:"response"`
}

func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	s := &Store{dir: dir, cases: map[string]*domain.TreatmentCase{}, events: map[string][]audit.Event{}, idem: map[string]Idempotent{}}
	s.load()
	return s, nil
}
func (s *Store) load() {
	entries, _ := os.ReadDir(s.dir)
	for _, e := range entries {
		if e.Name() == "idempotency.json" {
			if b, err := os.ReadFile(filepath.Join(s.dir, e.Name())); err == nil {
				json.Unmarshal(b, &s.idem)
			}
			continue
		}
		if filepath.Ext(e.Name()) == ".jsonl" {
			id := e.Name()[:len(e.Name())-len(".events.jsonl")]
			f, err := os.Open(filepath.Join(s.dir, e.Name()))
			if err == nil {
				sc := bufio.NewScanner(f)
				for sc.Scan() {
					var ev audit.Event
					if json.Unmarshal(sc.Bytes(), &ev) == nil {
						s.events[id] = append(s.events[id], ev)
					}
				}
				f.Close()
			}
			continue
		}
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			continue
		}
		var c domain.TreatmentCase
		if json.Unmarshal(b, &c) == nil {
			s.cases[c.CaseID] = &c
		}
	}
}
func (s *Store) Get(id string) (*domain.TreatmentCase, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.cases[id]
	if !ok {
		return nil, false
	}
	b, _ := json.Marshal(c)
	var cp domain.TreatmentCase
	json.Unmarshal(b, &cp)
	return &cp, true
}
func (s *Store) Events(id string) []audit.Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]audit.Event(nil), s.events[id]...)
}
func (s *Store) Save(c *domain.TreatmentCase, event audit.Event) error {
	s.mu.Lock()
	s.pending = &saveTransaction{
		caseID:      c.CaseID,
		caseValue:   c,
		nextEvents:  append(append([]audit.Event(nil), s.events[c.CaseID]...), event),
		tmp:         filepath.Join(s.dir, c.CaseID+".snapshot.tmp"),
		final:       filepath.Join(s.dir, c.CaseID+".json"),
		eventsTmp:   filepath.Join(s.dir, c.CaseID+".events.tmp"),
		eventsFinal: filepath.Join(s.dir, c.CaseID+".events.jsonl"),
	}
	s.mu.Unlock()
	if err := audit.Verify(s.pending.nextEvents); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s.pending.caseValue, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(s.pending.tmp, b, 0644); err != nil {
		return err
	}
	eventBytes := make([]byte, 0)
	for _, item := range s.pending.nextEvents {
		line, marshalErr := json.Marshal(item)
		if marshalErr != nil {
			_ = os.Remove(s.pending.tmp)
			return marshalErr
		}
		eventBytes = append(eventBytes, line...)
		eventBytes = append(eventBytes, '\n')
	}
	if err = os.WriteFile(s.pending.eventsTmp, eventBytes, 0644); err != nil {
		_ = os.Remove(s.pending.tmp)
		return err
	}
	oldSnapshot, oldErr := os.ReadFile(s.pending.final)
	if err = os.Rename(s.pending.tmp, s.pending.final); err != nil {
		_ = os.Remove(s.pending.eventsTmp)
		return err
	}
	if err = os.Rename(s.pending.eventsTmp, s.pending.eventsFinal); err != nil {
		if oldErr == nil {
			_ = os.WriteFile(s.pending.final, oldSnapshot, 0644)
		} else {
			_ = os.Remove(s.pending.final)
		}
		return err
	}
	s.mu.Lock()
	s.cases[s.pending.caseID] = s.pending.caseValue
	s.events[s.pending.caseID] = s.pending.nextEvents
	s.mu.Unlock()
	return nil
}
func (s *Store) SaveIdempotency(id, fp string, resp []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.idem[id] = Idempotent{fp, resp}
	b, _ := json.Marshal(s.idem)
	return os.WriteFile(filepath.Join(s.dir, "idempotency.json"), b, 0644)
}
func (s *Store) GetIdempotency(id string) (Idempotent, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.idem[id]
	return v, ok
}
func (s *Store) Validate() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, es := range s.events {
		if err := audit.Verify(es); err != nil {
			return fmt.Errorf("%s: %w", id, err)
		}
	}
	return nil
}
