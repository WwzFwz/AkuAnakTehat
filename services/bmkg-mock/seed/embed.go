package seed

import (
	"embed"
	"encoding/json"
	"example.com/akuanaktehat/bmkg-mock/internal/domain"
	"example.com/akuanaktehat/bmkg-mock/internal/store"
	"fmt"
)

//go:embed *.json
var fixtures embed.FS

func Load(data *store.Memory) error {
	var events []domain.SeismicEvent
	var warnings []domain.TsunamiWarning
	raw, err := fixtures.ReadFile("seismic-events.json")
	if err != nil {
		return err
	}
	if err = json.Unmarshal(raw, &events); err != nil {
		return err
	}
	raw, err = fixtures.ReadFile("tsunami-warnings.json")
	if err != nil {
		return err
	}
	if err = json.Unmarshal(raw, &warnings); err != nil {
		return err
	}
	if len(events) < 20 {
		return fmt.Errorf("expected >=20 seed seismic events")
	}
	byID := map[string]domain.SeismicEvent{}
	for _, e := range events {
		byID[e.EventID] = e
		data.AddSeismic(e)
	}
	for _, w := range warnings {
		e, ok := byID[w.RelatedEventID]
		if !ok || !e.PotentialTsunami {
			return fmt.Errorf("invalid seed warning reference")
		}
		w.ModifiedAt = e.OccurredAt
		data.SaveWarning(w)
	}
	return nil
}
