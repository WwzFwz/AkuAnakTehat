package sourcehttp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	BaseURL, Header, Credential string
	HTTP                        *http.Client
	MaxBytes                    int64
}

func New(base, header, credential string, timeout time.Duration) *Client {
	return &Client{BaseURL: strings.TrimRight(base, "/"), Header: header, Credential: credential, MaxBytes: 8 << 20, HTTP: &http.Client{Timeout: timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }, Transport: &http.Transport{DialContext: (&net.Dialer{Timeout: timeout, KeepAlive: 30 * time.Second}).DialContext, MaxConnsPerHost: 2, MaxIdleConnsPerHost: 2, MaxIdleConns: 4, IdleConnTimeout: 60 * time.Second, ResponseHeaderTimeout: timeout}}}
}
func (c *Client) Get(ctx context.Context, path string, since time.Time, corr string) ([]byte, error) {
	address := c.BaseURL + path
	if !since.IsZero() {
		address += "?since=" + url.QueryEscape(since.UTC().Format(time.RFC3339Nano))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, errors.New("invalid_source_url")
	}
	req.Header.Set(c.Header, c.Credential)
	req.Header.Set("X-Correlation-ID", corr)
	res, err := c.HTTP.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		var timeout net.Error
		if errors.As(err, &timeout) && timeout.Timeout() {
			return nil, errors.New("source_timeout")
		}
		return nil, errors.New("source_unavailable")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("source_http_%d", res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, c.MaxBytes+1))
	if err != nil {
		return nil, errors.New("source_body_unavailable")
	}
	if int64(len(body)) > c.MaxBytes {
		return nil, errors.New("source_body_too_large")
	}
	return body, nil
}
func (c *Client) Close() { c.HTTP.CloseIdleConnections() }
