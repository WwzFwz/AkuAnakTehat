package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"example.com/akuanaktehat/aggregator/internal/application/query"
	"example.com/akuanaktehat/aggregator/internal/domain/hazard"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeQuery struct {
	listFilter query.HazardFilter
	listPage   query.HazardPage
	listErr    error
	getEvent   hazard.Event
	getErr     error
	listCalls  int
	getCalls   int
}

func (f *fakeQuery) List(_ context.Context, filter query.HazardFilter) (query.HazardPage, error) {
	f.listCalls++
	f.listFilter = filter
	return f.listPage, f.listErr
}

func (f *fakeQuery) Get(_ context.Context, _ string) (hazard.Event, error) {
	f.getCalls++
	return f.getEvent, f.getErr
}

func TestListSuccessAndCorrelation(t *testing.T) {
	fake := &fakeQuery{listPage: query.HazardPage{
		Data:    []hazard.Event{{ID: "00000000-0000-0000-0000-000000000001", Source: "BMKG", Type: "SEISMIC"}},
		Sources: []query.SourceStatus{{Source: "BMKG", Status: "HEALTHY"}},
	}}
	handler := New(fake, "internal-secret", time.Second)
	req := httptest.NewRequest(http.MethodGet, "/internal/hazards?type=SEISMIC&limit=2&since=2026-09-01T00%3A00%3A00Z", nil)
	req.Header.Set("X-Internal-Key", "internal-secret")
	req.Header.Set("X-Correlation-ID", "query-test")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d; want %d", recorder.Code, http.StatusOK)
	}
	if recorder.Header().Get("X-Correlation-ID") != "query-test" {
		t.Fatal("correlation ID was not returned")
	}
	var page query.HazardPage
	if err := json.Unmarshal(recorder.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Data) != 1 || len(page.Sources) != 1 {
		t.Fatalf("unexpected page: %+v", page)
	}
	if fake.listFilter.Type != "SEISMIC" || fake.listFilter.Limit != 2 || fake.listFilter.Since == nil {
		t.Fatalf("unexpected filter: %+v", fake.listFilter)
	}
}

func TestListRejectsCredentialsAndQuery(t *testing.T) {
	for _, tc := range []struct {
		name string
		key  string
		path string
		code string
	}{
		{name: "missing key", path: "/internal/hazards", code: "invalid_credentials"},
		{name: "wrong key", key: "wrong", path: "/internal/hazards", code: "invalid_credentials"},
		{name: "unknown parameter", key: "internal-secret", path: "/internal/hazards?unknown=x", code: "invalid_query"},
		{name: "invalid cursor", key: "internal-secret", path: "/internal/hazards?cursor=bad", code: "invalid_cursor"},
		{name: "zero limit", key: "internal-secret", path: "/internal/hazards?limit=0", code: "invalid_limit"},
		{name: "negative limit", key: "internal-secret", path: "/internal/hazards?limit=-1", code: "invalid_limit"},
		{name: "over limit", key: "internal-secret", path: "/internal/hazards?limit=501", code: "invalid_limit"},
		{name: "malformed escape", key: "internal-secret", path: "/internal/hazards?severity=%zz", code: "invalid_query"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeQuery{}
			handler := New(fake, "internal-secret", time.Second)
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.key != "" {
				req.Header.Set("X-Internal-Key", tc.key)
			}
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			wantStatus := http.StatusUnauthorized
			if tc.code != "invalid_credentials" {
				wantStatus = http.StatusBadRequest
			}
			if recorder.Code != wantStatus {
				t.Fatalf("status=%d; want %d", recorder.Code, wantStatus)
			}
			var response errorResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Error != tc.code || response.CorrelationID == "" {
				t.Fatalf("response=%+v; want code %s and correlation ID", response, tc.code)
			}
			if fake.listCalls != 0 {
				t.Fatal("invalid request reached query service")
			}
		})
	}
}

func TestDetailNotFound(t *testing.T) {
	fake := &fakeQuery{getErr: query.ErrNotFound}
	handler := New(fake, "internal-secret", time.Second)
	req := httptest.NewRequest(http.MethodGet, "/internal/hazards/00000000-0000-0000-0000-000000000099", nil)
	req.Header.Set("X-Internal-Key", "internal-secret")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status=%d; want %d", recorder.Code, http.StatusNotFound)
	}
	if fake.getCalls != 1 {
		t.Fatalf("get calls=%d; want 1", fake.getCalls)
	}
}

func TestQueryTimeoutIsUnavailable(t *testing.T) {
	fake := &fakeQuery{}
	fake.listErr = context.DeadlineExceeded
	handler := New(fake, "internal-secret", time.Millisecond)
	req := httptest.NewRequest(http.MethodGet, "/internal/hazards", nil)
	req.Header.Set("X-Internal-Key", "internal-secret")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d; want %d", recorder.Code, http.StatusServiceUnavailable)
	}
	var response errorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Error != "dependency_unavailable" {
		t.Fatalf("error=%q; want dependency_unavailable", response.Error)
	}
	if !errors.Is(fake.listErr, context.DeadlineExceeded) {
		t.Fatal("test setup did not use deadline error")
	}
}
