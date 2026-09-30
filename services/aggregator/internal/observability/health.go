package observability

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

func Handler(ping func(context.Context) error, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	reply := func(w http.ResponseWriter, status int, value string) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]string{"status": value})
	}
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { reply(w, 200, "ok") })
	mux.HandleFunc("GET /ready/ingest", func(w http.ResponseWriter, r *http.Request) {
		if ping(r.Context()) != nil {
			reply(w, 503, "database_unavailable")
			return
		}
		reply(w, 200, "ready")
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) { reply(w, 503, "query_not_implemented") })
	unavailable := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(503)
		json.NewEncoder(w).Encode(map[string]string{"error": "query_not_implemented"})
	}
	mux.HandleFunc("GET /internal/hazards", unavailable)
	mux.HandleFunc("GET /internal/hazards/{rest...}", unavailable)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		mux.ServeHTTP(w, r)
		logger.Debug("http_request", "method", r.Method, "path", r.URL.Path, "latency_ms", time.Since(start).Milliseconds())
	})
}
