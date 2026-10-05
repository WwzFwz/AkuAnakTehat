package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

func Handler(ingestPing, queryPing func(context.Context) error, internal http.Handler, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	reply := func(w http.ResponseWriter, status int, value string) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]string{"status": value})
	}
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { reply(w, 200, "ok") })
	mux.HandleFunc("GET /ready/ingest", func(w http.ResponseWriter, r *http.Request) {
		if ingestPing(r.Context()) != nil {
			reply(w, 503, "database_unavailable")
			return
		}
		reply(w, 200, "ready")
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		if queryPing(r.Context()) != nil {
			reply(w, 503, "query_unavailable")
			return
		}
		reply(w, 200, "ready")
	})
	if internal != nil {
		mux.Handle("/internal/", internal)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := ensureCorrelation(w, r)
		start := time.Now()
		mux.ServeHTTP(w, r)
		logger.Debug("http_request", "method", r.Method, "path", r.URL.Path, "correlation_id", correlationID, "latency_ms", time.Since(start).Milliseconds())
	})
}

func ensureCorrelation(w http.ResponseWriter, r *http.Request) string {
	id := r.Header.Get("X-Correlation-ID")
	if len(id) == 0 || len(id) > 128 {
		id = ""
	}
	for _, char := range id {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '.' || char == '_' || char == '-') {
			id = ""
			break
		}
	}
	if id == "" {
		bytes := make([]byte, 16)
		if _, err := rand.Read(bytes); err == nil {
			id = hex.EncodeToString(bytes)
		} else {
			id = "aggregator-unknown"
		}
	}
	r.Header.Set("X-Correlation-ID", id)
	w.Header().Set("X-Correlation-ID", id)
	return id
}
