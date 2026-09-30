package store

import (
	"example.com/akuanaktehat/bmkg-mock/internal/domain"
	"sync"
	"time"
)

type Memory struct {
	mu       sync.RWMutex
	events   []domain.SeismicEvent
	warnings []domain.TsunamiWarning
}

func New() *Memory { return &Memory{} }
func (m *Memory) AddSeismic(e domain.SeismicEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, e)
}
func clone(w domain.TsunamiWarning) domain.TsunamiWarning {
	w.AffectedZones = append([]string{}, w.AffectedZones...)
	return w
}
func (m *Memory) SaveWarning(w domain.TsunamiWarning) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.warnings {
		if m.warnings[i].WarningID == w.WarningID {
			m.warnings[i] = clone(w)
			return
		}
	}
	m.warnings = append(m.warnings, clone(w))
}
func (m *Memory) ListSeismicSince(since time.Time) []domain.SeismicEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []domain.SeismicEvent{}
	for _, e := range m.events {
		if !e.OccurredAt.Before(since) {
			out = append(out, e)
		}
	}
	return out
}
func (m *Memory) ListWarningsSince(since time.Time) []domain.TsunamiWarning {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []domain.TsunamiWarning{}
	for _, w := range m.warnings {
		if !w.ModifiedAt.Before(since) {
			out = append(out, clone(w))
		}
	}
	return out
}
