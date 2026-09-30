package httpapi

import (
	"encoding/json"
	"example.com/akuanaktehat/pvmbg-mock/internal/auth"
	"example.com/akuanaktehat/pvmbg-mock/internal/config"
	"example.com/akuanaktehat/pvmbg-mock/internal/simulation"
	"example.com/akuanaktehat/pvmbg-mock/internal/store"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"
)

func New(cfg config.Config, data *store.Memory, state *simulation.State, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"status": "ok", "simulation": state.Snapshot()})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		s := state.Snapshot()
		code := 200
		if s.Outage {
			code = 503
		}
		writeJSON(w, code, map[string]any{"ready": !s.Outage, "simulation": s})
	})
	mux.HandleFunc("GET /volcanic-reports", func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || !strings.HasPrefix(token, "pvmbg_") || len(token) <= 6 || !auth.Matches(token, cfg.TokenHash) {
			failure(w, 401, "invalid PVMBG credentials")
			return
		}
		since, err := since(r)
		if err != nil {
			failure(w, 400, "since must be an RFC3339 timestamp")
			return
		}
		if !available(w, r, state) {
			return
		}
		duration := cfg.DelayMin
		if cfg.DelayMax > cfg.DelayMin {
			duration += time.Duration(rand.Int64N(int64(cfg.DelayMax-cfg.DelayMin) + 1))
		}
		// Admin changes wake the delay immediately; an enabled outage also affects in-flight data requests.
		timer := time.NewTimer(duration)
		defer timer.Stop()
		for {
			s, changed := state.Watch()
			if s.Outage {
				if !available(w, r, state) {
					return
				}
			}
			select {
			case <-r.Context().Done():
				return
			case <-changed:
				continue
			case <-timer.C:
				if !available(w, r, state) {
					return
				}
				writeJSON(w, 200, data.ListReportsSince(since))
				return
			}
		}
	})
	admin := func(fn http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !auth.Matches(r.Header.Get("X-Admin-Key"), cfg.AdminHash) {
				failure(w, 401, "invalid admin credentials")
				return
			}
			fn(w, r)
		}
	}
	mux.HandleFunc("POST /admin/outage", admin(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Enabled *bool  `json:"enabled"`
			Mode    string `json:"mode"`
		}
		if !decode(w, r, &body) {
			return
		}
		if body.Mode != "" && body.Mode != "hang" && body.Mode != "error" {
			failure(w, 400, "mode must be hang or error")
			return
		}
		writeJSON(w, 200, state.Outage(body.Enabled, body.Mode))
	}))
	mux.HandleFunc("POST /admin/schema-version", admin(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Version *int  `json:"version"`
			Enabled *bool `json:"enabled"`
		}
		if !decode(w, r, &body) {
			return
		}
		if body.Version != nil && body.Enabled != nil {
			failure(w, 400, "use either version or enabled")
			return
		}
		version := 0
		if body.Version != nil {
			version = *body.Version
			if version != 1 && version != 2 {
				failure(w, 400, "version must be 1 or 2")
				return
			}
		} else if body.Enabled != nil {
			version = 1
			if *body.Enabled {
				version = 2
			}
		}
		writeJSON(w, 200, state.Schema(version))
	}))
	return observe(mux, logger)
}

func available(w http.ResponseWriter, r *http.Request, state *simulation.State) bool {
	for {
		s, changed := state.Watch()
		if !s.Outage {
			return true
		}
		if s.Mode == "error" {
			failure(w, 503, "simulated PVMBG outage")
			return false
		}
		select {
		case <-r.Context().Done():
			return false
		case <-changed:
		}
	}
}
func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(dst); err != nil {
		if err == io.EOF {
			return true
		}
		failure(w, 400, "invalid JSON body")
		return false
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		failure(w, 400, "body must contain one JSON object")
		return false
	}
	return true
}
