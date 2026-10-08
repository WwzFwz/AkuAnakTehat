package aggregator

import (
	"bytes"
	"context"
	"encoding/json"
	"example.com/akuanaktehat/client-api/internal/application"
	"example.com/akuanaktehat/client-api/internal/observability"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	BaseURL, Key string
	Timeout      time.Duration
	HTTP         *http.Client
}

func New(base, key string, timeout time.Duration) *Client {
	return &Client{BaseURL: strings.TrimRight(base, "/"), Key: key, Timeout: timeout, HTTP: &http.Client{Timeout: timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }, Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, DialContext: (&net.Dialer{Timeout: timeout, KeepAlive: 30 * time.Second}).DialContext, MaxIdleConns: 100, MaxIdleConnsPerHost: 100, MaxConnsPerHost: 100, IdleConnTimeout: 60 * time.Second, ResponseHeaderTimeout: timeout}}}
}
func (c *Client) fetch(ctx context.Context, path string, out any) error {
	start := time.Now()
	defer func() {
		observability.Logger().Info("upstream_request", "operation", "aggregator_get", "correlation_id", observability.ID(ctx), "latency_ms", time.Since(start).Milliseconds())
	}()
	if c.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.Timeout)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+path, nil)
	if err != nil {
		return application.ErrUnavailable
	}
	req.Header.Set("X-Internal-Key", c.Key)
	req.Header.Set("X-Correlation-ID", observability.ID(ctx))
	resp, err := c.HTTP.Do(req)
	if err != nil && ctx.Err() == nil {
		resp, err = c.HTTP.Do(req.Clone(ctx))
	}
	if err != nil {
		return application.ErrUnavailable
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, (8<<20)+1))
	if err != nil || len(b) > 8<<20 {
		return application.ErrUnavailable
	}
	if resp.StatusCode != http.StatusOK {
		return classifyUpstream(resp.StatusCode, b)
	}
	if out != nil {
		decoder := json.NewDecoder(bytes.NewReader(b))
		decoder.UseNumber()
		if decoder.Decode(out) != nil {
			return application.ErrUnavailable
		}
		var extra any
		if decoder.Decode(&extra) != io.EOF {
			return application.ErrUnavailable
		}
	}
	return nil
}

func classifyUpstream(status int, body []byte) error {
	switch status {
	case http.StatusBadRequest:
		var response struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &response) != nil {
			return application.ErrQuery
		}
		switch response.Error {
		case "invalid_type":
			return application.ErrInvalidType
		case "invalid_severity":
			return application.ErrInvalidSeverity
		case "invalid_since":
			return application.ErrInvalidSince
		case "invalid_limit":
			return application.ErrInvalidLimit
		case "invalid_cursor":
			return application.ErrInvalidCursor
		default:
			return application.ErrQuery
		}
	case http.StatusNotFound:
		return application.ErrNotFound
	case http.StatusTooManyRequests:
		return application.ErrOverloaded
	default:
		return application.ErrUnavailable
	}
}
func (c *Client) List(ctx context.Context, q url.Values) (application.Page, error) {
	var p application.Page
	err := c.fetch(ctx, "/internal/hazards?"+q.Encode(), &p)
	if err == nil && (p.Data == nil || p.Sources == nil) {
		err = application.ErrUnavailable
	}
	return p, err
}
func (c *Client) Get(ctx context.Context, id string) (map[string]any, error) {
	var h map[string]any
	err := c.fetch(ctx, "/internal/hazards/"+url.PathEscape(id), &h)
	if err == nil && (h == nil || h["hazard_id"] == nil) {
		err = application.ErrUnavailable
	}
	return h, err
}
func (c *Client) Ready(ctx context.Context) error { return c.fetch(ctx, "/ready", nil) }
