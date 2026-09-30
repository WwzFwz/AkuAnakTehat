package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"time"
)

type contextKey struct{}

var validID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

func ID(ctx context.Context) string { v, _ := ctx.Value(contextKey{}).(string); return v }
func Logger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "client-api")
}

type response struct {
	http.ResponseWriter
	status int
}

func (w *response) WriteHeader(s int) { w.status = s; w.ResponseWriter.WriteHeader(s) }
func Middleware(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Correlation-ID")
		if !validID.MatchString(id) {
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		w.Header().Set("X-Correlation-ID", id)
		r = r.WithContext(context.WithValue(r.Context(), contextKey{}, id))
		start := time.Now()
		rw := &response{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)
		log.Info("request", "correlation_id", id, "operation", r.Method, "latency_ms", time.Since(start).Milliseconds(), "status", rw.status)
	})
}
