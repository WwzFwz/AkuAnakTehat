package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"time"
)

var correlationPattern = regexp.MustCompile(`^[a-zA-Z0-9._:-]{1,128}$`)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func failure(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
func since(r *http.Request) (time.Time, error) {
	value := r.URL.Query().Get("since")
	if value == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339Nano, value)
}
func delay(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(code int) {
	if r.status == 0 {
		r.status = code
		r.ResponseWriter.WriteHeader(code)
	}
}
func (r *responseRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(data)
}
func observe(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Correlation-ID")
		if !correlationPattern.MatchString(id) {
			var b [16]byte
			if _, err := rand.Read(b[:]); err != nil {
				failure(w, 503, "random source unavailable")
				return
			}
			id = hex.EncodeToString(b[:])
		}
		w.Header().Set("X-Correlation-ID", id)
		start := time.Now()
		recorder := &responseRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		status := recorder.status
		if status == 0 {
			status = 200
			if r.Context().Err() != nil {
				status = 499
			}
		}
		logger.Info("http_request", "correlation_id", id, "method", r.Method, "path", r.URL.Path, "status", status, "latency_ms", float64(time.Since(start).Microseconds())/1000)
	})
}
