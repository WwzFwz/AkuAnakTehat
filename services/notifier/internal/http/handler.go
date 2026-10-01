package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type Health interface{ Ping(context.Context) error }
type View interface {
	List(context.Context, string, int) ([]json.RawMessage, error)
}

func Handler(db Health, kafka Health, view View) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { write(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if db.Ping(ctx) != nil || kafka.Ping(ctx) != nil {
			write(w, 503, map[string]string{"error": "dependency_unavailable"})
			return
		}
		write(w, 200, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("GET /processed", func(w http.ResponseWriter, r *http.Request) {
		limit := 100
		if raw := r.URL.Query().Get("limit"); raw != "" {
			v, err := strconv.Atoi(raw)
			if err != nil || v < 1 || v > 200 {
				write(w, 400, map[string]string{"error": "invalid_limit"})
				return
			}
			limit = v
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		data, err := view.List(ctx, r.URL.Query().Get("after"), limit+1)
		if err != nil {
			write(w, 503, map[string]string{"error": "store_unavailable"})
			return
		}
		next := ""
		if len(data) > limit {
			data = data[:limit]
			var last struct {
				ID string `json:"event_id"`
			}
			_ = json.Unmarshal(data[len(data)-1], &last)
			next = last.ID
		}
		write(w, 200, map[string]any{"data": data, "next_cursor": next})
	})
	return mux
}
func write(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
