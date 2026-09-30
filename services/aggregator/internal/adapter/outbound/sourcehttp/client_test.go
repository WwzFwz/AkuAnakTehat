package sourcehttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDeadlineCredentialsAndResponseLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-BMKG-Key") != "test-key" || r.Header.Get("X-Correlation-ID") != "test-correlation" {
			t.Error("credential or correlation header missing")
		}
		if r.URL.Query().Get("since") != "2026-09-01T00:00:00Z" {
			t.Error("since not encoded")
		}
		if r.URL.Path == "/hang" {
			<-r.Context().Done()
			return
		}
		w.Write([]byte("123456789"))
	}))
	defer server.Close()
	c := New(server.URL, "X-BMKG-Key", "test-key", 50*time.Millisecond)
	defer c.Close()
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	start := time.Now()
	if _, err := c.Get(context.Background(), "/hang", since, "test-correlation"); err == nil {
		t.Fatal("expected deadline")
	}
	if time.Since(start) > time.Second {
		t.Fatal("deadline not bounded")
	}
	c.MaxBytes = 4
	if _, err := c.Get(context.Background(), "/big", since, "test-correlation"); err == nil {
		t.Fatal("oversized response accepted")
	}
}
