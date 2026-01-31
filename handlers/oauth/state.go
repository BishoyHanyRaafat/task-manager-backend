package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

type StateData struct {
	Platform string
	Mode     string // "login" | "link"
	UserID   string // used for Mode="link"
}

type stateEntry struct {
	expiresAt time.Time
	data      StateData
}

type StateStore struct {
	mu    sync.Mutex
	store map[string]stateEntry
	ttl   time.Duration
}

func NewStateStore(ttl time.Duration) *StateStore {
	s := &StateStore{
		store: make(map[string]stateEntry),
		ttl:   ttl,
	}
	go s.cleanupLoop(10 * time.Minute)
	return s
}

func (s *StateStore) Generate() string {
	return s.GenerateWithData(StateData{})
}

func (s *StateStore) GenerateWithData(data StateData) string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	s.mu.Lock()
	s.store[state] = stateEntry{
		expiresAt: time.Now().Add(s.ttl),
		data:      data,
	}
	s.mu.Unlock()
	return state
}

func (s *StateStore) Validate(state string) bool {
	_, ok := s.Consume(state)
	return ok
}

// Consume validates the state token and returns any data stored with it.
// It is one-time use: a successful call deletes the token from the store.
func (s *StateStore) Consume(state string) (StateData, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.store[state]
	if !ok {
		return StateData{}, false
	}
	delete(s.store, state) // one-time use
	if time.Now().After(entry.expiresAt) {
		return StateData{}, false
	}
	return entry.data, true
}

func (s *StateStore) cleanupLoop(every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for range t.C {
		now := time.Now()
		s.mu.Lock()
		for k, exp := range s.store {
			if now.After(exp.expiresAt) {
				delete(s.store, k)
			}
		}
		s.mu.Unlock()
	}
}
