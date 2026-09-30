package reference

import (
	"embed"
	"encoding/json"
	"example.com/akuanaktehat/aggregator/internal/application/canonicalize"
)

//go:embed volcanoes.json
var files embed.FS

func Load() (map[string]canonicalize.Volcano, error) {
	b, err := files.ReadFile("volcanoes.json")
	if err != nil {
		return nil, err
	}
	var m map[string]canonicalize.Volcano
	err = json.Unmarshal(b, &m)
	return m, err
}
