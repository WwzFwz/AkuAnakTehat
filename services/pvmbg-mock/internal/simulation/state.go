package simulation

import "sync"

type Snapshot struct {
	Outage        bool   `json:"outage"`
	Mode          string `json:"mode"`
	SchemaVersion int    `json:"schema_version"`
}
type State struct {
	mu      sync.RWMutex
	value   Snapshot
	changed chan struct{}
}

func New() *State {
	return &State{value: Snapshot{Mode: "error", SchemaVersion: 1}, changed: make(chan struct{})}
}
func (s *State) Snapshot() Snapshot { s.mu.RLock(); defer s.mu.RUnlock(); return s.value }
func (s *State) Watch() (Snapshot, <-chan struct{}) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value, s.changed
}
func (s *State) Outage(enabled *bool, mode string) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	if enabled == nil {
		s.value.Outage = !s.value.Outage
	} else {
		s.value.Outage = *enabled
	}
	if mode != "" {
		s.value.Mode = mode
	}
	close(s.changed)
	s.changed = make(chan struct{})
	return s.value
}
func (s *State) Schema(version int) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	if version == 0 {
		if s.value.SchemaVersion == 1 {
			version = 2
		} else {
			version = 1
		}
	}
	s.value.SchemaVersion = version
	return s.value
}
