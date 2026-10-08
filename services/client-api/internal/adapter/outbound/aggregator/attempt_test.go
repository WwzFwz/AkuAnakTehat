package aggregator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"example.com/akuanaktehat/client-api/internal/application"
	"example.com/akuanaktehat/client-api/internal/observability"
)

type attemptTransport func(*http.Request) (*http.Response, error)

func (f attemptTransport) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

type attemptLog struct {
	Message       string   `json:"msg"`
	Operation     string   `json:"operation"`
	CorrelationID string   `json:"correlation_id"`
	Attempt       int      `json:"attempt"`
	LatencyMS     *float64 `json:"latency_ms"`
	Status        int      `json:"status"`
	Result        string   `json:"result"`
}

func readAttemptLogs(t *testing.T, data []byte) []attemptLog {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(data))
	var records []attemptLog
	for {
		var record attemptLog
		if err := decoder.Decode(&record); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		if record.Message != "upstream_request" || record.Operation != "aggregator_get" || record.LatencyMS == nil || *record.LatencyMS < 0 {
			t.Fatalf("invalid attempt log: %+v", record)
		}
		records = append(records, record)
	}
	return records
}

func TestRetryLogsEachAttempt(t *testing.T) {
	var output bytes.Buffer
	client := New("http://aggregator.test", "test-only-internal-key", time.Second)
	client.logger = slog.New(slog.NewJSONHandler(&output, nil))
	const correlationID = "retry-correlation-test"
	var firstDeadline time.Time
	attempts := 0
	client.HTTP.Transport = attemptTransport(func(req *http.Request) (*http.Response, error) {
		attempts++
		if req.Header.Get("X-Correlation-ID") != correlationID || req.Header.Get("X-Internal-Key") != client.Key {
			t.Fatal("required headers not preserved across retry")
		}
		deadline, ok := req.Context().Deadline()
		if !ok {
			t.Fatal("missing request deadline")
		}
		if attempts == 1 {
			firstDeadline = deadline
			return nil, errors.New("transport error with test-only-internal-key")
		}
		if !deadline.Equal(firstDeadline) {
			t.Fatal("retry extended the shared timeout budget")
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"hazard_id":"test-hazard"}`)), Header: make(http.Header)}, nil
	})
	// Use the actual inbound middleware to attach the correlation ID to context.
	handler := observability.Middleware(slog.New(slog.NewJSONHandler(io.Discard, nil)), http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		hazard, err := client.Get(req.Context(), "test-hazard?private-query-value")
		if err != nil || hazard["hazard_id"] != "test-hazard" {
			t.Fatalf("retry did not recover: hazard=%v err=%v", hazard, err)
		}
	}))
	req := httptest.NewRequest(http.MethodGet, "/hazards", nil)
	req.Header.Set("X-Correlation-ID", correlationID)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	records := readAttemptLogs(t, output.Bytes())
	if attempts != 2 || len(records) != 2 {
		t.Fatalf("calls=%d logs=%d; want two of each", attempts, len(records))
	}
	for i, record := range records {
		if record.Attempt != i+1 || record.CorrelationID != correlationID {
			t.Fatalf("attempt identity lost: %+v", record)
		}
	}
	if records[0].Result != "transport_error" || records[0].Status != 0 || records[1].Result != "success" || records[1].Status != 200 {
		t.Fatalf("incorrect outcomes: %+v", records)
	}
	for _, sensitive := range []string{client.Key, "private-query-value", "aggregator.test", "transport error with"} {
		if strings.Contains(output.String(), sensitive) {
			t.Fatalf("attempt log exposed %q", sensitive)
		}
	}
}

func TestAttemptFailuresPreserveRetryPolicy(t *testing.T) {
	for _, tc := range []struct {
		name       string
		status     int
		body       string
		cancel     bool
		transport  bool
		wantCount  int
		wantResult string
		wantError  error
	}{
		{name: "transport exhausted", transport: true, wantCount: 2, wantResult: "transport_error", wantError: application.ErrUnavailable},
		{name: "canceled", cancel: true, wantCount: 1, wantResult: "canceled", wantError: application.ErrUnavailable},
		{name: "http unavailable", status: 503, body: "private-response-body", wantCount: 1, wantResult: "http_error", wantError: application.ErrUnavailable},
		{name: "http not found", status: 404, body: "private-response-body", wantCount: 1, wantResult: "http_error", wantError: application.ErrNotFound},
		{name: "invalid json", status: 200, body: "private-response-body", wantCount: 1, wantResult: "invalid_json", wantError: application.ErrUnavailable},
		{name: "trailing json", status: 200, body: `{} {}`, wantCount: 1, wantResult: "invalid_json", wantError: application.ErrUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			client := New("http://aggregator.test", "test-only-key", time.Second)
			client.logger = slog.New(slog.NewJSONHandler(&output, nil))
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls := 0
			client.HTTP.Transport = attemptTransport(func(req *http.Request) (*http.Response, error) {
				calls++
				if tc.cancel {
					cancel()
					return nil, req.Context().Err()
				}
				if tc.transport {
					return nil, errors.New("private-transport-error")
				}
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body)), Header: make(http.Header)}, nil
			})
			_, err := client.Get(ctx, "test")
			if !errors.Is(err, tc.wantError) {
				t.Fatalf("error=%v, want %v", err, tc.wantError)
			}
			records := readAttemptLogs(t, output.Bytes())
			if calls != tc.wantCount || len(records) != tc.wantCount {
				t.Fatalf("calls=%d logs=%d, want %d", calls, len(records), tc.wantCount)
			}
			for i, record := range records {
				if record.Attempt != i+1 || record.Result != tc.wantResult || record.Status != tc.status {
					t.Fatalf("incorrect attempt log: %+v", record)
				}
			}
			if strings.Contains(output.String(), "private-") {
				t.Fatal("log exposed upstream error or response body")
			}
		})
	}
}

type observedBody struct {
	io.Reader
	closed bool
}

func (b *observedBody) Close() error { b.closed = true; return nil }

func TestAttemptLogsTimeoutAndClosesBody(t *testing.T) {
	t.Run("timeout stops retry", func(t *testing.T) {
		var output bytes.Buffer
		client := New("http://aggregator.test", "key", 20*time.Millisecond)
		client.logger = slog.New(slog.NewJSONHandler(&output, nil))
		client.HTTP.Transport = attemptTransport(func(req *http.Request) (*http.Response, error) {
			<-req.Context().Done()
			return nil, req.Context().Err()
		})
		if err := client.Ready(context.Background()); !errors.Is(err, application.ErrUnavailable) {
			t.Fatalf("error=%v", err)
		}
		records := readAttemptLogs(t, output.Bytes())
		if len(records) != 1 || records[0].Result != "timeout" {
			t.Fatalf("incorrect timeout log: %+v", records)
		}
	})
	t.Run("body consumed and closed", func(t *testing.T) {
		var output bytes.Buffer
		client := New("http://aggregator.test", "key", time.Second)
		client.logger = slog.New(slog.NewJSONHandler(&output, nil))
		body := &observedBody{Reader: strings.NewReader(`{"hazard_id":"test"}`)}
		client.HTTP.Transport = attemptTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: body, Header: make(http.Header)}, nil
		})
		if _, err := client.Get(context.Background(), "test"); err != nil {
			t.Fatal(err)
		}
		if !body.closed {
			t.Fatal("response body not closed")
		}
		records := readAttemptLogs(t, output.Bytes())
		if len(records) != 1 || records[0].Attempt != 1 || records[0].Result != "success" {
			t.Fatalf("incorrect success log: %+v", records)
		}
	})
}
