package store

import (
	"example.com/akuanaktehat/pvmbg-mock/internal/domain"
	"sync"
	"time"
)

type Memory struct {
	mu      sync.RWMutex
	reports []domain.VolcanicReport
}

func New() *Memory { return &Memory{} }
func clone(r domain.VolcanicReport) domain.VolcanicReport {
	if r.ConfidenceLevel != nil {
		v := *r.ConfidenceLevel
		r.ConfidenceLevel = &v
	}
	return r
}
func (m *Memory) AddReport(r domain.VolcanicReport) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reports = append(m.reports, clone(r))
}
func (m *Memory) ListReportsSince(since time.Time) []domain.VolcanicReport {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []domain.VolcanicReport{}
	for _, r := range m.reports {
		if !r.ReportedAt.Before(since) {
			out = append(out, clone(r))
		}
	}
	return out
}
