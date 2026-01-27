package httpx

import (
	"context"
	"net/http"
	"time"
)

type JavaHTTPClient struct {
	BaseURL string
	Client  *http.Client
}

func NewJavaHTTPClient(baseURL string) *JavaHTTPClient {
	return &JavaHTTPClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (c *JavaHTTPClient) Do(ctx context.Context, method, path, traceID string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Trace-Id", traceID)
	req.Header.Set("traceparent", "00-"+traceID+"-0000000000000000-01")
	return c.Client.Do(req)
}
