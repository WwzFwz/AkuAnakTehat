package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultTimeout = 5 * time.Second
	refreshSkew    = 5 * time.Second
	maxBodyBytes   = 4 << 20
)

var (
	ErrInvalidConfiguration = errors.New("invalid CLI configuration")
	ErrInvalidResponse      = errors.New("invalid service response")
	ErrResponseTooLarge     = errors.New("service response is too large")
)

type APIError struct {
	Status int
	Code   string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("service request failed: %s", e.Code)
	}
	return fmt.Sprintf("service request failed with status %d", e.Status)
}

type TokenSource struct {
	client       *http.Client
	authURL      string
	clientID     string
	clientSecret string
	skew         time.Duration
	now          func() time.Time

	mu        sync.Mutex
	access    string
	refresh   string
	expiresAt time.Time
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type tokenError struct {
	Error string `json:"error"`
}

func NewTokenSource(httpClient *http.Client, authURL, clientID, clientSecret string) *TokenSource {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}
	return &TokenSource{
		client:       httpClient,
		authURL:      strings.TrimRight(authURL, "/"),
		clientID:     clientID,
		clientSecret: clientSecret,
		skew:         refreshSkew,
		now:          time.Now,
	}
}

// Token serializes refresh operations because the auth service rotates refresh
// tokens and invalidates the previous token immediately.
func (s *TokenSource) Token(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	if s.access != "" && now.Add(s.skew).Before(s.expiresAt) {
		return s.access, nil
	}

	if s.refresh != "" {
		pair, err := s.request(ctx, url.Values{
			"grant_type":    {"refresh_token"},
			"refresh_token": {s.refresh},
		})
		if err == nil {
			s.set(pair)
			return s.access, nil
		}
		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.Code != "invalid_grant" {
			return "", err
		}
	}

	if s.clientID == "" || s.clientSecret == "" {
		return "", ErrInvalidConfiguration
	}
	pair, err := s.request(ctx, url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {s.clientID},
		"client_secret": {s.clientSecret},
	})
	if err != nil {
		return "", err
	}
	s.set(pair)
	return s.access, nil
}

func (s *TokenSource) Invalidate() {
	s.mu.Lock()
	s.access = ""
	s.expiresAt = time.Time{}
	s.mu.Unlock()
}

func (s *TokenSource) set(pair tokenResponse) {
	s.access = pair.AccessToken
	s.refresh = pair.RefreshToken
	s.expiresAt = s.now().Add(time.Duration(pair.ExpiresIn) * time.Second)
}

func (s *TokenSource) request(ctx context.Context, values url.Values) (tokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.authURL+"/oauth/token", strings.NewReader(values.Encode()))
	if err != nil {
		return tokenResponse{}, ErrInvalidConfiguration
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.client.Do(req)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := readBody(resp.Body)
	if err != nil {
		return tokenResponse{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var failure tokenError
		_ = json.Unmarshal(body, &failure)
		return tokenResponse{}, &APIError{Status: resp.StatusCode, Code: failure.Error}
	}
	var pair tokenResponse
	if err := json.Unmarshal(body, &pair); err != nil || pair.AccessToken == "" || pair.RefreshToken == "" || pair.ExpiresIn <= 0 {
		return tokenResponse{}, ErrInvalidResponse
	}
	return pair, nil
}

type Config struct {
	APIURL       string
	AuthURL      string
	ClientID     string
	ClientSecret string
	Timeout      time.Duration
}

type Client struct {
	apiURL     *url.URL
	httpClient *http.Client
	tokens     *TokenSource
}

func New(cfg Config) (*Client, error) {
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}
	apiURL, err := parseBaseURL(cfg.APIURL)
	if err != nil {
		return nil, err
	}
	authURL, err := parseBaseURL(cfg.AuthURL)
	if err != nil {
		return nil, err
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	httpClient := &http.Client{Transport: transport, Timeout: cfg.Timeout}
	return &Client{
		apiURL:     apiURL,
		httpClient: httpClient,
		tokens:     NewTokenSource(httpClient, authURL.String(), cfg.ClientID, cfg.ClientSecret),
	}, nil
}

type ListOptions struct {
	Type     string
	Severity string
	Since    string
	Limit    int
	Cursor   string
	Raw      bool
}

func (c *Client) List(ctx context.Context, options ListOptions) ([]byte, error) {
	query, err := options.query()
	if err != nil {
		return nil, err
	}
	return c.do(ctx, "/v1/hazards", query)
}

func (c *Client) Get(ctx context.Context, id string, raw bool) ([]byte, error) {
	id = strings.TrimSpace(id)
	if id == "" || strings.ContainsAny(id, "/\\") {
		return nil, fmt.Errorf("invalid hazard id")
	}
	path := "/v1/hazards/" + id
	if raw {
		path += "/raw"
	}
	return c.do(ctx, path, nil)
}

func (o ListOptions) query() (url.Values, error) {
	query := make(url.Values)
	if o.Type != "" {
		o.Type = strings.ToUpper(o.Type)
		if o.Type != "SEISMIC" && o.Type != "VOLCANIC" {
			return nil, fmt.Errorf("type must be SEISMIC or VOLCANIC")
		}
		query.Set("type", o.Type)
	}
	if o.Severity != "" {
		o.Severity = strings.ToUpper(o.Severity)
		switch o.Severity {
		case "NORMAL", "WASPADA", "SIAGA", "AWAS":
		default:
			return nil, fmt.Errorf("invalid severity")
		}
		query.Set("severity", o.Severity)
	}
	if o.Since != "" {
		parsed, err := time.Parse(time.RFC3339, o.Since)
		if err != nil {
			return nil, fmt.Errorf("since must be RFC3339")
		}
		query.Set("since", parsed.UTC().Format(time.RFC3339Nano))
	}
	if o.Limit < 1 || o.Limit > 500 {
		return nil, fmt.Errorf("limit must be between 1 and 500")
	}
	query.Set("limit", strconv.Itoa(o.Limit))
	if o.Cursor != "" {
		query.Set("cursor", o.Cursor)
	}
	if o.Raw {
		query.Set("include", "raw")
	}
	return query, nil
}

func (c *Client) do(ctx context.Context, path string, query url.Values) ([]byte, error) {
	token, err := c.tokens.Token(ctx)
	if err != nil {
		return nil, err
	}
	correlationID := newCorrelationID()
	for attempt := 0; attempt < 2; attempt++ {
		u := *c.apiURL
		u.Path = strings.TrimRight(u.Path, "/") + path
		u.RawQuery = query.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, ErrInvalidConfiguration
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-Correlation-ID", correlationID)
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("API request failed: %w", err)
		}
		body, readErr := readBody(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		if resp.StatusCode == http.StatusUnauthorized && attempt == 0 {
			c.tokens.Invalidate()
			token, err = c.tokens.Token(ctx)
			if err != nil {
				return nil, err
			}
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			var failure tokenError
			_ = json.Unmarshal(body, &failure)
			return nil, &APIError{Status: resp.StatusCode, Code: failure.Error}
		}
		return body, nil
	}
	return nil, &APIError{Status: http.StatusUnauthorized, Code: "invalid_token"}
}

func PrintJSON(w io.Writer, body []byte) error {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, body, "", "  "); err != nil {
		return ErrInvalidResponse
	}
	pretty.WriteByte('\n')
	_, err := pretty.WriteTo(w)
	return err
}

func readBody(r io.Reader) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(r, maxBodyBytes+1))
	if err != nil {
		return nil, fmt.Errorf("reading service response: %w", err)
	}
	if len(body) > maxBodyBytes {
		return nil, ErrResponseTooLarge
	}
	return body, nil
}

func parseBaseURL(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimRight(strings.TrimSpace(raw), "/"))
	if err != nil || u.Scheme == "" || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, ErrInvalidConfiguration
	}
	u.RawQuery = ""
	u.Fragment = ""
	return u, nil
}

func newCorrelationID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err == nil {
		return hex.EncodeToString(b)
	}
	return "field-cli-" + strconv.FormatInt(time.Now().UnixNano(), 10)
}
