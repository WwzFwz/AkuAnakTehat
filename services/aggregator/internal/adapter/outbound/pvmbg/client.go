package pvmbg

import (
	"context"
	"example.com/akuanaktehat/aggregator/internal/adapter/outbound/sourcehttp"
	"example.com/akuanaktehat/aggregator/internal/application/canonicalize"
	"time"
)

type Client struct{ *sourcehttp.Client }

func New(base, token string, timeout time.Duration) *Client {
	return &Client{sourcehttp.New(base, "Authorization", "Bearer "+token, timeout)}
}
func (c *Client) FetchReports(ctx context.Context, since time.Time, corr string) ([]canonicalize.Item, error) {
	b, err := c.Get(ctx, "/volcanic-reports", since, corr)
	if err != nil {
		return nil, err
	}
	return canonicalize.Decode(canonicalize.VolcanicEndpoint, b)
}
