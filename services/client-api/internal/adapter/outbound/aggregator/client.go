package aggregator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"example.com/akuanaktehat/client-api/internal/application"
	"example.com/akuanaktehat/client-api/internal/observability"
	"io"
	"log/slog"
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
	logger       *slog.Logger
}

func New(base, key string, timeout time.Duration) *Client {
	return &Client{BaseURL: strings.TrimRight(base, "/"), Key: key, Timeout: timeout, HTTP: &http.Client{Timeout: timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }, Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, DialContext: (&net.Dialer{Timeout: timeout, KeepAlive: 30 * time.Second}).DialContext, MaxIdleConns: 100, MaxIdleConnsPerHost: 100, MaxConnsPerHost: 100, IdleConnTimeout: 60 * time.Second, ResponseHeaderTimeout: timeout}}}
}
func (c *Client) fetch(ctx context.Context, path string, out any) error {
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
	for attempt := 1; attempt <= 2; attempt++ {
		retry, err := c.fetchAttempt(req.Clone(ctx), attempt, out)
		if !retry || ctx.Err() != nil || attempt == 2 {
			return err
		}
	}
	return application.ErrUnavailable
}

// fetchAttempt measures one HTTP attempt through body consumption and decoding.
// Only transport failures may be retried; the caller owns the shared deadline.
func (c *Client) fetchAttempt(req *http.Request, attempt int, out any) (bool, error) {
	start := time.Now()
	status, result := 0, "success"
	defer func() {
		logger := c.logger
		if logger == nil {
			logger = observability.Logger()
		}
		logger.Info("upstream_request", "operation", "aggregator_get",
			"correlation_id", observability.ID(req.Context()), "attempt", attempt,
			"latency_ms", float64(time.Since(start))/float64(time.Millisecond),
			"status", status, "result", result)
	}()
	resp, err := c.HTTP.Do(req)
	if err != nil {
		result = "transport_error"
		if errors.Is(err, context.Canceled) {
			result = "canceled"
		} else if errors.Is(err, context.DeadlineExceeded) {
			result = "timeout"
		}
		return true, application.ErrUnavailable
	}
	status = resp.StatusCode
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, (8<<20)+1))
	if err != nil {
		result = "body_read_error"
		return false, application.ErrUnavailable
	}
	if len(b) > 8<<20 {
		result = "response_too_large"
		return false, application.ErrUnavailable
	}
	if resp.StatusCode != http.StatusOK {
		result = "http_error"
		return false, classifyUpstream(resp.StatusCode, b)
	}
	if out != nil {
		decoder := json.NewDecoder(bytes.NewReader(b))
		decoder.UseNumber()
		if decoder.Decode(out) != nil {
			result = "invalid_json"
			return false, application.ErrUnavailable
		}
		var extra any
		if decoder.Decode(&extra) != io.EOF {
			result = "invalid_json"
			return false, application.ErrUnavailable
		}
	}
	return false, nil
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
