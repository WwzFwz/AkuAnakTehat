package seed

import (
	"embed"
	"encoding/json"
	"example.com/akuanaktehat/pvmbg-mock/internal/domain"
	"example.com/akuanaktehat/pvmbg-mock/internal/store"
	"fmt"
)

//go:embed *.json
var fixtures embed.FS

func Load(data *store.Memory) error {
	var reports []domain.VolcanicReport
	raw, err := fixtures.ReadFile("volcanic-reports.json")
	if err != nil {
		return err
	}
	if err = json.Unmarshal(raw, &reports); err != nil {
		return err
	}
	if len(reports) < 20 {
		return fmt.Errorf("expected >=20 seed volcanic reports")
	}
	for _, r := range reports {
		if r.ConfidenceLevel != nil {
			return fmt.Errorf("seed must use schema v1")
		}
		data.AddReport(r)
	}
	return nil
}
