package bmkg

import (
	"context"
	"example.com/akuanaktehat/aggregator/internal/adapter/outbound/sourcehttp"
	"example.com/akuanaktehat/aggregator/internal/application/canonicalize"
	"time"
)

type Client struct{ *sourcehttp.Client }

func New(base, key string, timeout time.Duration) *Client {
	return &Client{sourcehttp.New(base, "X-BMKG-Key", key, timeout)}
}
func (c *Client) FetchSeismic(ctx context.Context, since time.Time, corr string) ([]canonicalize.Item, error) {
	b, err := c.Get(ctx, "/seismic-events", since, corr)
	if err != nil {
		return nil, err
	}
	return canonicalize.Decode(canonicalize.SeismicEndpoint, b)
}
func (c *Client) FetchWarnings(ctx context.Context, since time.Time, corr string) ([]canonicalize.Item, error) {
	b, err := c.Get(ctx, "/tsunami-warnings", since, corr)
	if err != nil {
		return nil, err
	}
	return canonicalize.Decode(canonicalize.WarningEndpoint, b)
}
