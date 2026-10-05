package httpapi

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"example.com/akuanaktehat/client-api/internal/application"
	"example.com/akuanaktehat/client-api/internal/authn"
	"example.com/akuanaktehat/client-api/internal/middleware"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestListValidationErrorCodes(t *testing.T) {
	for _, tc := range []struct {
		name string
		path string
		want string
	}{
		{name: "route type conflict", path: "/v1/hazards/seismic?type=VOLCANIC", want: "invalid_type"},
		{name: "severity", path: "/v1/hazards?severity=CRITICAL", want: "invalid_severity"},
		{name: "since", path: "/v1/hazards?since=invalid", want: "invalid_since"},
		{name: "cursor", path: "/v1/hazards?cursor=bad", want: "invalid_cursor"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pub, priv, err := ed25519.GenerateKey(rand.Reader)
			if err != nil {
				t.Fatal(err)
			}
			upstream := &validationUpstream{}
			handler := New(application.Service{Upstream: upstream}, authn.Verifier{Key: pub, Issuer: "test", Audience: "test"}, middleware.New(100, 100, 10), 100, 500)
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set("Authorization", "Bearer "+testToken(t, priv, "media"))
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status=%d; want %d", recorder.Code, http.StatusBadRequest)
			}
			if !strings.Contains(recorder.Body.String(), `"error":"`+tc.want+`"`) {
				t.Fatalf("body=%s; want error %s", recorder.Body.String(), tc.want)
			}
			if tc.want != "invalid_cursor" && upstream.calls != 0 {
				t.Fatal("invalid request reached upstream")
			}
		})
	}
}

func TestRouterDistinguishesConcurrencyRejection(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	upstream := &blockingUpstream{started: make(chan struct{}), release: make(chan struct{})}
	handler := New(application.Service{Upstream: upstream}, authn.Verifier{Key: pub, Issuer: "test", Audience: "test"}, middleware.New(100, 100, 1), 100, 500)

	firstDone := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		req := httptest.NewRequest(http.MethodGet, "/v1/hazards", nil)
		req.Header.Set("Authorization", "Bearer "+testToken(t, priv, "media"))
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		firstDone <- recorder
	}()
	<-upstream.started

	second := httptest.NewRecorder()
	secondRequest := httptest.NewRequest(http.MethodGet, "/v1/hazards", nil)
	secondRequest.Header.Set("Authorization", "Bearer "+testToken(t, priv, "media"))
	handler.ServeHTTP(second, secondRequest)
	if second.Code != http.StatusTooManyRequests || !strings.Contains(second.Body.String(), `"error":"upstream_overloaded"`) {
		t.Fatalf("status=%d body=%s; want upstream_overloaded", second.Code, second.Body.String())
	}
	close(upstream.release)
	if first := <-firstDone; first.Code != http.StatusOK {
		t.Fatalf("first request status=%d; want %d", first.Code, http.StatusOK)
	}
}

type validationUpstream struct{ calls int }

func (u *validationUpstream) List(_ context.Context, values url.Values) (application.Page, error) {
	// Cursor structure is opaque to client-api; Aggregator owns its validation.
	// The fake mirrors the upstream response for an invalid cursor.
	u.calls++
	if values.Get("cursor") == "bad" {
		return application.Page{}, application.ErrInvalidCursor
	}
	return application.Page{Data: []map[string]any{}, Sources: []application.Source{}}, nil
}

func (u *validationUpstream) Get(context.Context, string) (map[string]any, error) {
	u.calls++
	return map[string]any{"hazard_id": "demo"}, nil
}

func (u *validationUpstream) Ready(context.Context) error { return nil }

type blockingUpstream struct {
	started chan struct{}
	release chan struct{}
}

func (u *blockingUpstream) List(ctx context.Context, _ url.Values) (application.Page, error) {
	close(u.started)
	select {
	case <-u.release:
		return application.Page{Data: []map[string]any{}, Sources: []application.Source{}}, nil
	case <-ctx.Done():
		return application.Page{}, ctx.Err()
	}
}

func (u *blockingUpstream) Get(context.Context, string) (map[string]any, error) {
	return map[string]any{"hazard_id": "demo"}, nil
}

func (u *blockingUpstream) Ready(context.Context) error { return nil }

func testToken(t *testing.T, private ed25519.PrivateKey, subject string) string {
	t.Helper()
	now := time.Now()
	claims := authn.Claims{Scope: "hazard:read:summary", RegisteredClaims: jwt.RegisteredClaims{
		Issuer: "test", Audience: jwt.ClaimStrings{"test"}, Subject: subject, ID: "router-test", IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
	}}
	token, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(private)
	if err != nil {
		t.Fatal(err)
	}
	return token
}
