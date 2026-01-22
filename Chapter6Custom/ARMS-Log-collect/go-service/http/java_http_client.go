package httpx

import (
	"context"
	"fmt"
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
		Client:  &http.Client{Timeout: 3 * time.Second},
	}
}

func (c *JavaHTTPClient) Do(ctx context.Context, method, path, traceID string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, nil)
	if err != nil {
		return nil, err
	}

	// trace 透传
	req.Header.Set("X-Trace-Id", traceID)

	req.Header.Set("X-Caller-Service", "go-service")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("java http error: status=%d path=%s", resp.StatusCode, path)
	}

	return resp, nil
}
