package httpapi

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"example.com/akuanaktehat/client-api/internal/application"
	"example.com/akuanaktehat/client-api/internal/authn"
	"example.com/akuanaktehat/client-api/internal/middleware"
	"github.com/golang-jwt/jwt/v5"
)

type upstream struct{ calls int }

func (u *upstream) Get(context.Context, string) (map[string]any, error) {
	u.calls++
	return map[string]any{"hazard_id": "demo", "source": "BMKG", "hazard_type": "SEISMIC", "severity": "SIAGA", "area_name": "Demo", "occurred_at": "2026-09-01T00:00:00Z", "ingested_at": "2026-09-01T00:00:01Z", "source_ref_id": "raw-id", "latitude": -7.5, "longitude": 110.4, "attributes": map[string]any{"magnitude": 6.7}, "future_internal_field": "must-not-leak", "sources": []map[string]any{{"source": "BMKG", "status": "DEGRADED", "stale_since": "2026-09-01T00:01:00Z", "private_metadata": "must-not-leak"}}}, nil
}
func (u *upstream) List(ctx context.Context, _ url.Values) (application.Page, error) {
	h, _ := u.Get(ctx, "demo")
	return application.Page{Data: []map[string]any{h}, Sources: []application.Source{{Source: "BMKG", Status: "HEALTHY"}}}, nil
}
func (u *upstream) Ready(context.Context) error { return nil }

func TestHTTPProjectionAndForbiddenRequests(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, scope, path     string
		status, fields, calls int
	}{
		{"media detail", "hazard:read:summary", "/v1/hazards/demo", 200, 7, 1},
		{"media list", "hazard:read:summary", "/v1/hazards", 200, 7, 1},
		{"media selection", "hazard:read:summary", "/v1/hazards/demo?fields=hazard_id,severity", 200, 2, 1},
		{"media raw route", "hazard:read:summary", "/v1/hazards/demo/raw", 403, 0, 0},
		{"media raw include", "hazard:read:summary", "/v1/hazards?include=raw", 403, 0, 0},
		{"media raw fields", "hazard:read:summary", "/v1/hazards?fields=hazard_id,attributes", 403, 0, 0},
		{"unknown field", "hazard:read:summary", "/v1/hazards?fields=future_internal_field", 400, 0, 0},
		{"privileged raw", "hazard:read:summary hazard:read:raw", "/v1/hazards/demo/raw", 200, 11, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			u := &upstream{}
			h := New(application.Service{Upstream: u}, authn.Verifier{Key: pub, Issuer: "test", Audience: "test"}, middleware.New(100, 100, 10), 100, 500)
			claims := authn.Claims{Scope: tc.scope, RegisteredClaims: jwt.RegisteredClaims{Issuer: "test", Audience: jwt.ClaimStrings{"test"}, Subject: "test", ID: "unique", IssuedAt: jwt.NewNumericDate(time.Now()), ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))}}
			token, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(priv)
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			if w.Code != tc.status {
				t.Fatalf("status=%d;want %d", w.Code, tc.status)
			}
			if u.calls != tc.calls {
				t.Fatalf("upstream calls=%d;want %d", u.calls, tc.calls)
			}
			if tc.status != 200 {
				return
			}
			var body map[string]any
			if json.Unmarshal(w.Body.Bytes(), &body) != nil {
				t.Fatal("invalid JSON")
			}
			if rows, ok := body["data"].([]any); ok {
				if len(rows) != 1 {
					t.Fatal("expected one row")
				}
				body = rows[0].(map[string]any)
			} else {
				sources, ok := body["sources"].([]any)
				if !ok || len(sources) != 1 {
					t.Fatal("detail lost freshness metadata")
				}
				source := sources[0].(map[string]any)
				if source["status"] != "DEGRADED" || source["stale_since"] == nil || len(source) != 3 {
					t.Fatal("freshness metadata missing or leaking upstream fields")
				}
			}
			delete(body, "sources") // envelope metadata is not a HazardEvent field
			if len(body) != tc.fields {
				t.Fatalf("returned %d fields;want %d", len(body), tc.fields)
			}
			if _, ok := body["future_internal_field"]; ok {
				t.Fatal("new upstream field escaped allowlist")
			}
			if tc.scope == "hazard:read:summary" {
				for _, key := range []string{"source_ref_id", "latitude", "longitude", "attributes"} {
					if _, ok := body[key]; ok {
						t.Fatalf("Media received %s", key)
					}
				}
			}
		})
	}
}
