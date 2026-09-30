package httpapi

import (
	"encoding/json"
	"errors"
	"example.com/akuanaktehat/client-api/internal/application"
	"example.com/akuanaktehat/client-api/internal/authn"
	"example.com/akuanaktehat/client-api/internal/authz"
	"example.com/akuanaktehat/client-api/internal/middleware"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func New(app application.Service, verifier authn.Verifier, limits *middleware.Limits, pageDefault, pageMax int) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { write(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		if app.Upstream.Ready(r.Context()) != nil {
			write(w, 503, map[string]string{"status": "unavailable"})
			return
		}
		write(w, 200, map[string]string{"status": "ready"})
	})
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		header := strings.Fields(r.Header.Get("Authorization"))
		if len(header) != 2 || !strings.EqualFold(header[0], "Bearer") {
			write(w, 401, map[string]string{"error": "invalid_token"})
			return
		}
		claims, err := verifier.Verify(header[1])
		if err != nil {
			write(w, 401, map[string]string{"error": "invalid_token"})
			return
		}
		release, ok := limits.Enter(claims.Subject)
		if !ok {
			w.Header().Set("Retry-After", "1")
			write(w, 429, map[string]string{"error": "rate_limited"})
			return
		}
		defer release()
		q, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			write(w, 400, map[string]string{"error": "invalid_query"})
			return
		}
		for key, values := range q {
			if len(values) != 1 {
				write(w, 400, map[string]string{"error": "invalid_query"})
				return
			}
			switch key {
			case "fields", "include", "limit", "cursor", "type", "severity", "since":
			default:
				write(w, 400, map[string]string{"error": "invalid_query"})
				return
			}
		}
		var fields []string
		if q.Has("fields") {
			fields = strings.Split(q.Get("fields"), ",")
			for i := range fields {
				fields[i] = strings.TrimSpace(fields[i])
			}
		}
		raw := strings.HasSuffix(r.URL.Path, "/raw")
		if q.Has("include") {
			if q.Get("include") != "raw" {
				write(w, 400, map[string]string{"error": "invalid_include"})
				return
			}
			raw = true
		}
		if err = authz.Allowed(claims.Scope, fields, raw); err != nil {
			respondError(w, err)
			return
		}
		q.Del("fields")
		q.Del("include")
		path := strings.TrimPrefix(r.URL.Path, "/v1/hazards")
		var response any
		if path == "" || path == "/seismic" || path == "/volcanic" {
			if path == "/seismic" {
				q.Set("type", "SEISMIC")
			}
			if path == "/volcanic" {
				q.Set("type", "VOLCANIC")
			}
			if typ := q.Get("type"); typ != "" && typ != "SEISMIC" && typ != "VOLCANIC" {
				write(w, 400, map[string]string{"error": "invalid_type"})
				return
			}
			limit := pageDefault
			if severity := q.Get("severity"); severity != "" && severity != "NORMAL" && severity != "WASPADA" && severity != "SIAGA" && severity != "AWAS" {
				write(w, 400, map[string]string{"error": "invalid_query"})
				return
			}
			if since := q.Get("since"); since != "" {
				parsed, err := time.Parse(time.RFC3339, since)
				if err != nil {
					write(w, 400, map[string]string{"error": "invalid_query"})
					return
				}
				q.Set("since", parsed.UTC().Format(time.RFC3339Nano))
			}
			if q.Has("limit") {
				limit, err = strconv.Atoi(q.Get("limit"))
				if err != nil || limit < 1 || limit > pageMax {
					write(w, 400, map[string]string{"error": "invalid_limit"})
					return
				}
			}
			if len(q.Get("cursor")) > 2048 {
				write(w, 400, map[string]string{"error": "invalid_cursor"})
				return
			}
			q.Set("limit", strconv.Itoa(limit))
			response, err = app.List(r.Context(), claims.Scope, q, fields, raw)
		} else {
			id := strings.TrimPrefix(strings.TrimSuffix(path, "/raw"), "/")
			if id == "" || strings.Contains(id, "/") || len(id) > 128 {
				write(w, 404, map[string]string{"error": "not_found"})
				return
			}
			response, err = app.Get(r.Context(), claims.Scope, id, fields, raw)
		}
		if err != nil {
			respondError(w, err)
			return
		}
		write(w, 200, response)
	}
	mux.HandleFunc("GET /v1/hazards", handler)
	mux.HandleFunc("GET /v1/hazards/{rest...}", handler)
	return mux
}
func respondError(w http.ResponseWriter, err error) {
	status, code := 503, "dependency_unavailable"
	switch {
	case errors.Is(err, authz.ErrForbidden):
		status, code = 403, "insufficient_scope"
	case errors.Is(err, authz.ErrFields):
		status, code = 400, "invalid_fields"
	case errors.Is(err, application.ErrNotFound):
		status, code = 404, "not_found"
	case errors.Is(err, application.ErrQuery):
		status, code = 400, "invalid_query"
	case errors.Is(err, application.ErrOverloaded):
		status, code = 429, "upstream_overloaded"
		w.Header().Set("Retry-After", "1")
	}
	write(w, status, map[string]string{"error": code})
}
