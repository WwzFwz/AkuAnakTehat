package generator

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"example.com/akuanaktehat/bmkg-mock/internal/domain"
	"example.com/akuanaktehat/bmkg-mock/internal/store"
	"fmt"
	"log/slog"
	"time"
)

type Generator struct {
	prefix  string
	tick    uint64
	late    *domain.SeismicEvent
	early   *domain.SeismicEvent
	warning *domain.TsunamiWarning
}

func New() (*Generator, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return nil, err
	}
	return &Generator{prefix: hex.EncodeToString(b[:])}, nil
}

// Tick is called serially by Run. It exercises warning-before-event, warning-after-event and escalation.
func (g *Generator) Tick(data *store.Memory, now time.Time) {
	g.tick++
	if g.warning != nil {
		w := *g.warning
		w.ThreatLevel = "Awas"
		w.ModifiedAt = now
		data.SaveWarning(w)
		g.warning = nil
	}
	if g.late != nil {
		w := warning(*g.late, now)
		data.SaveWarning(w)
		g.warning = &w
		g.late = nil
	}
	if g.early != nil {
		e := *g.early
		e.OccurredAt = now
		data.AddSeismic(e)
		g.early = nil
	}
	e := domain.SeismicEvent{EventID: fmt.Sprintf("bmkg-%s-%d", g.prefix, g.tick), Magnitude: 4.5 + float64(g.tick%4), DepthKM: 15, EpicenterLat: -7.5, EpicenterLon: 110.2, RegionName: "Wilayah Demo Sintetis", OccurredAt: now, PotentialTsunami: g.tick%3 == 0}
	data.AddSeismic(e)
	if e.PotentialTsunami {
		g.late = &e
	}
	if g.tick%4 == 0 {
		future := e
		future.EventID += "-early"
		future.PotentialTsunami = true
		data.SaveWarning(warning(future, now))
		g.early = &future
	}
}
func warning(e domain.SeismicEvent, now time.Time) domain.TsunamiWarning {
	return domain.TsunamiWarning{WarningID: "warning-" + e.EventID, RelatedEventID: e.EventID, ThreatLevel: "Waspada", AffectedZones: []string{"Pesisir Demo Sintetis"}, EstimatedArrival: now.Add(30 * time.Minute), ModifiedAt: now}
}
func Run(ctx context.Context, data *store.Memory, interval time.Duration, logger *slog.Logger) {
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
			g.Tick(data, now.UTC())
			logger.Info("record_generated", "sequence", g.tick)
		}
	}
}
