package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"example.com/akuanaktehat/auth-service/internal/application"
	"mime"
	"net/http"
	"sync"
	"time"
)

func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func New(s *application.Service, rate, burst int) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { write(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		if s.Store.Ping(r.Context()) != nil {
			write(w, 503, map[string]string{"status": "unavailable"})
			return
		}
		write(w, 200, map[string]string{"status": "ready"})
	})
	var mu sync.Mutex
	tokens := float64(burst)
	last := time.Now()
	slots := make(chan struct{}, 100)
	mux.HandleFunc("POST /oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		mu.Lock()
		now := time.Now()
		tokens = min(float64(burst), tokens+now.Sub(last).Seconds()*float64(rate))
		last = now
		allowed := tokens >= 1
		if allowed {
			tokens--
		}
		mu.Unlock()
		if !allowed {
			w.Header().Set("Retry-After", "1")
			write(w, 429, map[string]string{"error": "rate_limited"})
			return
		}
		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
		default:
			w.Header().Set("Retry-After", "1")
			write(w, 429, map[string]string{"error": "overloaded"})
			return
		}
		media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || media != "application/x-www-form-urlencoded" {
			write(w, 415, map[string]string{"error": "invalid_request"})
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 8192)
		if err = r.ParseForm(); err != nil {
			var limit *http.MaxBytesError
			status := 400
			if errors.As(err, &limit) {
				status = 413
			}
			write(w, status, map[string]string{"error": "invalid_request"})
			return
		}
		for _, v := range r.PostForm {
			if len(v) != 1 {
				write(w, 400, map[string]string{"error": "invalid_request"})
				return
			}
		}
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		var pair any
		switch r.PostForm.Get("grant_type") {
		case "client_credentials":
			pair, err = s.Issue(ctx, r.PostForm.Get("client_id"), r.PostForm.Get("client_secret"))
		case "refresh_token":
			pair, err = s.Refresh(ctx, r.PostForm.Get("refresh_token"))
		default:
			write(w, 400, map[string]string{"error": "unsupported_grant_type"})
			return
		}
		if err != nil {
			status := 503
			code := "temporarily_unavailable"
			if errors.Is(err, application.ErrCredentials) {
				status = 401
				code = "invalid_client"
			}
			if errors.Is(err, application.ErrGrant) {
				status = 400
				code = "invalid_grant"
			}
			write(w, status, map[string]string{"error": code})
			return
		}
		write(w, 200, pair)
	})
	return mux
}
