package generator

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"example.com/akuanaktehat/pvmbg-mock/internal/domain"
	"example.com/akuanaktehat/pvmbg-mock/internal/simulation"
	"example.com/akuanaktehat/pvmbg-mock/internal/store"
	"fmt"
	"log/slog"
	"time"
)

type Generator struct {
	prefix string
	tick   uint64
}

func New() (*Generator, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return nil, err
	}
	return &Generator{prefix: hex.EncodeToString(b[:])}, nil
}
func (g *Generator) Tick(data *store.Memory, state *simulation.State, now time.Time) {
	g.tick++
	r := domain.VolcanicReport{ReportID: fmt.Sprintf("pvmbg-%s-%d", g.prefix, g.tick), VolcanoID: domain.VolcanoIDs[g.tick%2], AlertLevel: []string{"Normal", "Waspada", "Siaga", "Awas"}[g.tick%4], EruptionCount24H: int(g.tick % 10), AshColumnHeightM: float64(g.tick%5) * 500, ReportedAt: now}
	if state.Snapshot().SchemaVersion == 2 {
		confidence := 0.9
		r.ConfidenceLevel = &confidence
	}
	data.AddReport(r)
}
func Run(ctx context.Context, data *store.Memory, state *simulation.State, interval time.Duration, logger *slog.Logger) {
	g, err := New()
	if err != nil {
		logger.Error("generator_failed", "error", err)
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			g.Tick(data, state, now.UTC())
			logger.Info("record_generated", "sequence", g.tick, "schema_version", state.Snapshot().SchemaVersion)
		}
	}
}
