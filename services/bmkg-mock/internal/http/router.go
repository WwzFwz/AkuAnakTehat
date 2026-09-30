package httpapi

import (
	"example.com/akuanaktehat/bmkg-mock/internal/auth"
	"example.com/akuanaktehat/bmkg-mock/internal/config"
	"example.com/akuanaktehat/bmkg-mock/internal/store"
	"log/slog"
	"net/http"
)

func New(cfg config.Config, data *store.Memory, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	health := func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, map[string]string{"status": "ok"}) }
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /ready", health)
	for _, path := range []string{"/seismic-events", "/tsunami-warnings"} {
		mux.HandleFunc("GET "+path, func(w http.ResponseWriter, r *http.Request) {
			if !auth.Matches(r.Header.Get("X-BMKG-Key"), cfg.KeyHash) {
				failure(w, 401, "invalid BMKG credentials")
				return
			}
			since, err := since(r)
			if err != nil {
				failure(w, 400, "since must be an RFC3339 timestamp")
				return
			}
			if !delay(r.Context(), cfg.FixedDelay) {
				return
			}
			if r.URL.Path == "/seismic-events" {
				writeJSON(w, 200, data.ListSeismicSince(since))
			} else {
				writeJSON(w, 200, data.ListWarningsSince(since))
			}
		})
	}
	return observe(mux, logger)
}
