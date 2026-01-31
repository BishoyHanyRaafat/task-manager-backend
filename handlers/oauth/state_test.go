package oauth

import (
	"testing"
	"time"
)

func TestStateStore_ConsumeReturnsDataAndIsOneTime(t *testing.T) {
	s := NewStateStore(50 * time.Millisecond)
	state := s.GenerateWithData(StateData{Platform: "mobile"})

	got, ok := s.Consume(state)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if got.Platform != "mobile" {
		t.Fatalf("expected platform=mobile, got %q", got.Platform)
	}

	_, ok = s.Consume(state)
	if ok {
		t.Fatalf("expected ok=false on second consume (one-time use)")
	}
}

func TestStateStore_ConsumeExpiredReturnsFalse(t *testing.T) {
	s := NewStateStore(10 * time.Millisecond)
	state := s.GenerateWithData(StateData{Platform: "web"})

	time.Sleep(25 * time.Millisecond)
	_, ok := s.Consume(state)
	if ok {
		t.Fatalf("expected ok=false for expired state")
	}
}
