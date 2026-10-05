package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientRefreshesAfterUnauthorized(t *testing.T) {
	var issue, refresh, apiCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			w.Header().Set("Content-Type", "application/json")
			if r.Form.Get("grant_type") == "client_credentials" {
				issue.Add(1)
				_, _ = w.Write([]byte(`{"access_token":"access-1","token_type":"Bearer","expires_in":60,"refresh_token":"refresh-1"}`))
				return
			}
			refresh.Add(1)
			_, _ = w.Write([]byte(`{"access_token":"access-2","token_type":"Bearer","expires_in":60,"refresh_token":"refresh-2"}`))
		case "/v1/hazards":
			apiCalls.Add(1)
			if r.Header.Get("Authorization") == "Bearer access-1" {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"invalid_token"}`))
				return
			}
			_, _ = w.Write([]byte(`{"data":[],"sources":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	c, err := New(Config{APIURL: server.URL, AuthURL: server.URL, ClientID: "field-team", ClientSecret: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	body, err := c.List(context.Background(), ListOptions{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != `{"data":[],"sources":[]}` {
		t.Fatalf("unexpected response: %s", body)
	}
	if issue.Load() != 1 || refresh.Load() != 1 || apiCalls.Load() != 2 {
		t.Fatalf("issue=%d refresh=%d api_calls=%d", issue.Load(), refresh.Load(), apiCalls.Load())
	}
}

func TestRefreshIsSerialized(t *testing.T) {
	var refresh atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			w.Header().Set("Content-Type", "application/json")
			if r.Form.Get("grant_type") == "client_credentials" {
				_, _ = w.Write([]byte(`{"access_token":"access-1","token_type":"Bearer","expires_in":60,"refresh_token":"refresh-1"}`))
				return
			}
			refresh.Add(1)
			time.Sleep(20 * time.Millisecond)
			_, _ = w.Write([]byte(`{"access_token":"access-2","token_type":"Bearer","expires_in":60,"refresh_token":"refresh-2"}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[],"sources":[]}`))
	}))
	defer server.Close()

	c, err := New(Config{APIURL: server.URL, AuthURL: server.URL, ClientID: "field-team", ClientSecret: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = c.tokens.Token(context.Background()); err != nil {
		t.Fatal(err)
	}
	c.tokens.Invalidate()

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := c.List(context.Background(), ListOptions{Limit: 1}); err != nil {
				t.Errorf("list failed: %v", err)
			}
		}()
	}
	wg.Wait()
	if refresh.Load() != 1 {
		t.Fatalf("refresh requests=%d, want 1", refresh.Load())
	}
}

func TestListOptionsEncodeContract(t *testing.T) {
	query, err := (ListOptions{Type: "volcanic", Severity: "siaga", Since: "2026-01-02T03:04:05+07:00", Limit: 20, Cursor: "opaque", Raw: true}).query()
	if err != nil {
		t.Fatal(err)
	}
	want := url.Values{
		"type":     {"VOLCANIC"},
		"severity": {"SIAGA"},
		"since":    {"2026-01-01T20:04:05Z"},
		"limit":    {strconv.Itoa(20)},
		"cursor":   {"opaque"},
		"include":  {"raw"},
	}
	if query.Encode() != want.Encode() {
		t.Fatalf("query=%s want=%s", query.Encode(), want.Encode())
	}
}

func TestPrintJSON(t *testing.T) {
	var output []byte
	writer := writerFunc(func(p []byte) (int, error) {
		output = append(output, p...)
		return len(p), nil
	})
	if err := PrintJSON(writer, []byte(`{"data":[]}`)); err != nil {
		t.Fatal(err)
	}
	var value any
	if err := json.Unmarshal(output, &value); err != nil {
		t.Fatalf("output is not JSON: %v", err)
	}
}

type writerFunc func([]byte) (int, error)

func (w writerFunc) Write(p []byte) (int, error) { return w(p) }
