package aggregator

import (
	"context"
	"errors"
	"example.com/akuanaktehat/client-api/internal/application"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestClassifyUpstreamErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{name: "type", status: http.StatusBadRequest, body: `{"error":"invalid_type"}`, want: application.ErrInvalidType},
		{name: "severity", status: http.StatusBadRequest, body: `{"error":"invalid_severity"}`, want: application.ErrInvalidSeverity},
		{name: "since", status: http.StatusBadRequest, body: `{"error":"invalid_since"}`, want: application.ErrInvalidSince},
		{name: "limit", status: http.StatusBadRequest, body: `{"error":"invalid_limit"}`, want: application.ErrInvalidLimit},
		{name: "cursor", status: http.StatusBadRequest, body: `{"error":"invalid_cursor"}`, want: application.ErrInvalidCursor},
		{name: "not found", status: http.StatusNotFound, body: `{"error":"not_found"}`, want: application.ErrNotFound},
		{name: "overloaded", status: http.StatusTooManyRequests, body: `{"error":"upstream_overloaded"}`, want: application.ErrOverloaded},
		{name: "unavailable", status: http.StatusBadGateway, body: `{"error":"dependency_unavailable"}`, want: application.ErrUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifyUpstream(tc.status, []byte(tc.body)); !errors.Is(got, tc.want) {
				t.Fatalf("error=%v; want %v", got, tc.want)
			}
		})
	}
}

func TestClientListValidatesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Internal-Key") != "internal-secret" {
			t.Error("internal key was not forwarded")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[],"next_cursor":"next","sources":[]}`))
	}))
	defer server.Close()

	client := New(server.URL, "internal-secret", time.Second)
	page, err := client.List(context.Background(), mapValues("type", "SEISMIC"))
	if err != nil {
		t.Fatal(err)
	}
	if page.NextCursor != "next" || page.Data == nil || page.Sources == nil {
		t.Fatalf("unexpected page: %+v", page)
	}
}

func mapValues(key, value string) url.Values {
	return url.Values{key: []string{value}}
}
