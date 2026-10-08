package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"example.com/akuanaktehat/aggregator/internal/application/query"
	"example.com/akuanaktehat/aggregator/internal/observability"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type correlationKey struct{}

type handler struct {
	query       query.HazardQuery
	internalKey string
	timeout     time.Duration
}

func New(service query.HazardQuery, internalKey string, timeout time.Duration) http.Handler {
	if timeout <= 0 {
		timeout = time.Second
	}
	h := &handler{query: service, internalKey: internalKey, timeout: timeout}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /internal/hazards", h.list)
	mux.HandleFunc("GET /internal/hazards/{id}", h.detail)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r, id := withCorrelation(w, r)
		mux.ServeHTTP(w, r.WithContext(observability.WithID(context.WithValue(r.Context(), correlationKey{}, id), id)))
	})
}

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	if !validInternalKey(h.internalKey, r.Header.Get("X-Internal-Key")) {
		writeError(w, r, http.StatusUnauthorized, "invalid_credentials", "internal credential is invalid")
		return
	}
	if h.query == nil {
		writeMappedError(w, r, query.ErrUnavailable)
		return
	}
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		writeMappedError(w, r, query.ErrInvalidQuery)
		return
	}
	filter, err := parseFilter(values)
	if err != nil {
		writeMappedError(w, r, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()
	page, err := h.query.List(ctx, filter)
	if err != nil {
		writeMappedError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *handler) detail(w http.ResponseWriter, r *http.Request) {
	if !validInternalKey(h.internalKey, r.Header.Get("X-Internal-Key")) {
		writeError(w, r, http.StatusUnauthorized, "invalid_credentials", "internal credential is invalid")
		return
	}
	if r.URL.RawQuery != "" {
		writeMappedError(w, r, query.ErrInvalidQuery)
		return
	}
	if h.query == nil {
		writeMappedError(w, r, query.ErrUnavailable)
		return
	}
	id := r.PathValue("id")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, r, http.StatusNotFound, "not_found", "hazard was not found")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()
	event, err := h.query.Get(ctx, id)
	if err != nil {
		writeMappedError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, event)
}

func writeMappedError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := classifyError(err)
	writeError(w, r, status, code, message)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	writeJSON(w, status, errorResponse{Error: code, Message: message, CorrelationID: correlationID(r)})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func withCorrelation(w http.ResponseWriter, r *http.Request) (*http.Request, string) {
	id := r.Header.Get("X-Correlation-ID")
	if !validCorrelationID(id) {
		id = newCorrelationID()
	}
	r.Header.Set("X-Correlation-ID", id)
	w.Header().Set("X-Correlation-ID", id)
	return r, id
}

func correlationID(r *http.Request) string {
	if id, ok := r.Context().Value(correlationKey{}).(string); ok && id != "" {
		return id
	}
	return r.Header.Get("X-Correlation-ID")
}

func validCorrelationID(id string) bool {
	if len(id) == 0 || len(id) > 128 {
		return false
	}
	for _, r := range id {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-') {
			return false
		}
	}
	return true
}

func newCorrelationID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err == nil {
		return hex.EncodeToString(b)
	}
	return "aggregator-" + time.Now().UTC().Format("20060102150405.000000000")
}
