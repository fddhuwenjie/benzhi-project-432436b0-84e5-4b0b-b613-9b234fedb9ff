package audit

import (
	"testing"
	"time"
)

func TestChain(t *testing.T) {
	e := Append(nil, "c", 1, "x", "a", nil, time.Now())
	if err := Verify([]Event{e}); err != nil {
		t.Fatal(err)
	}
	e2 := Append([]Event{e}, "c", 2, "y", "a", nil, time.Now())
	if err := Verify([]Event{e, e2}); err != nil {
		t.Fatal(err)
	}
}
