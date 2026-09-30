package aggregator

import (
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
	HTTP         *http.Client
}

func New(base, key string, timeout time.Duration) *Client {
	return &Client{BaseURL: strings.TrimRight(base, "/"), Key: key, HTTP: &http.Client{Timeout: timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }, Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, DialContext: (&net.Dialer{Timeout: timeout, KeepAlive: 30 * time.Second}).DialContext, MaxIdleConns: 100, MaxIdleConnsPerHost: 100, MaxConnsPerHost: 100, IdleConnTimeout: 60 * time.Second, ResponseHeaderTimeout: timeout}}}
}
func (c *Client) fetch(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+path, nil)
	if err != nil {
		return application.ErrUnavailable
	}
	req.Header.Set("X-Internal-Key", c.Key)
	req.Header.Set("X-Correlation-ID", observability.ID(ctx))
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return application.ErrUnavailable
	}
	defer resp.Body.Close()
	if resp.StatusCode == 429 {
		return application.ErrOverloaded
	}
	if resp.StatusCode == 400 {
		return application.ErrQuery
	}
	if resp.StatusCode == 404 {
		return application.ErrNotFound
	}
	if resp.StatusCode != 200 {
		return application.ErrUnavailable
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, (4<<20)+1))
	if err != nil || len(b) > 4<<20 {
		return application.ErrUnavailable
	}
	if out != nil && json.Unmarshal(b, out) != nil {
		return application.ErrUnavailable
	}
	return nil
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
