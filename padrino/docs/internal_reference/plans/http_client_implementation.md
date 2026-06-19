# Plan: Resilient HTTP Client Implementation

This plan outlines the implementation of a resilient HTTP client wrapper in `pkg/httpclient`, following the project's hexagonal architecture and Go standards.

## 🎯 Objectives
- Implement a resilient `http.Client` wrapper with retry logic (backoff + jitter).
- Provide a custom `Response` struct with unexported fields and exported methods.
- Ensure proper context propagation and timeout management.
- Define consumer-side interfaces in `internal/core/usecases`.
- Implement adapters in `internal/adapters/externalapi`.

## 🛠️ Implementation Details

### 1. `pkg/httpclient`

#### `Response` Struct (`pkg/httpclient/response.go`)
```go
package httpclient

import "net/http"

// Response wraps the HTTP response data with unexported fields.
type Response struct {
	statusCode int
	body       []byte
	headers    http.Header
}

// IsSuccess returns true if the status code is 2xx.
func (r *Response) IsSuccess() bool {
	return r.statusCode >= 200 && r.statusCode < 300
}

// IsServerError returns true if the status code is 5xx.
func (r *Response) IsServerError() bool {
	return r.statusCode >= 500 && r.statusCode < 600
}

// IsClientError returns true if the status code is 4xx.
func (r *Response) IsClientError() bool {
	return r.statusCode >= 400 && r.statusCode < 500
}

// StatusCode returns the HTTP status code.
func (r *Response) StatusCode() int {
	return r.statusCode
}

// Body returns the response body bytes.
func (r *Response) Body() []byte {
	return r.body
}

// Headers returns the response headers.
func (r *Response) Headers() http.Header {
	return r.headers
}
```

#### `Client` Struct and Logic (`pkg/httpclient/client.go`)
```go
package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

// Client is a resilient HTTP client.
type Client struct {
	httpClient       *http.Client
	retryTimes       int
	retryStatusCodes []int
	minBackoff       time.Duration
	maxBackoff       time.Duration
}

// Option defines a functional configuration for the Client.
type Option func(*Client)

// New creates a new resilient Client with default settings and optional overrides.
func New(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		retryTimes: 3,
		retryStatusCodes: []int{
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		},
		minBackoff: 100 * time.Millisecond,
		maxBackoff: 2 * time.Second,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithRetryTimes sets the number of retries.
func WithRetryTimes(n int) Option {
	return func(c *Client) {
		c.retryTimes = n
	}
}

// WithRetryStatusCodes sets which status codes trigger a retry.
func WithRetryStatusCodes(codes []int) Option {
	return func(c *Client) {
		c.retryStatusCodes = codes
	}
}

// WithTimeout sets the request timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// Do executes the request with retry logic and context propagation.
func (c *Client) Do(ctx context.Context, req *http.Request) (*Response, error) {
	var lastResp *Response
	var lastErr error

	// Ensure we can retry the body if present
	if req.Body != nil && req.GetBody == nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(bodyBytes)), nil
		}
	}

	for i := 0; i <= c.retryTimes; i++ {
		if i > 0 {
			if err := c.backoff(ctx, i); err != nil {
				return nil, err
			}
			
			// Reset body for retry
			if req.GetBody != nil {
				newBody, err := req.GetBody()
				if err != nil {
					return nil, fmt.Errorf("failed to reset request body: %w", err)
				}
				req.Body = newBody
			}
		}

		resp, err := c.httpClient.Do(req.WithContext(ctx))
		if err != nil {
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		lastResp = &Response{
			statusCode: resp.StatusCode,
			body:       body,
			headers:    resp.Header,
		}

		if !c.shouldRetry(resp.StatusCode) {
			return lastResp, nil
		}
	}

	return lastResp, lastErr
}

func (c *Client) shouldRetry(statusCode int) bool {
	for _, code := range c.retryStatusCodes {
		if code == statusCode {
			return true
		}
	}
	return false
}

func (c *Client) backoff(ctx context.Context, attempt int) error {
	backoff := c.minBackoff * (1 << uint(attempt-1))
	if backoff > c.maxBackoff {
		backoff = c.maxBackoff
	}
	
	jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
	totalBackoff := backoff + jitter

	timer := time.NewTimer(totalBackoff)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
```

### 2. Consumer-Side Interface (`internal/core/usecases/ports.go`)

```go
package usecases

import (
	"context"
	"github.com/your-username/don/pkg/httpclient"
)

// HTTPClient defines the behavior expected by use cases.
// Following the "Consumer-Side Interface" standard.
type HTTPClient interface {
	Do(ctx context.Context, req *http.Request) (*httpclient.Response, error)
}
```

### 3. Adapter Implementation (`internal/adapters/externalapi/client.go`)

```go
package externalapi

import (
	"context"
	"net/http"
	"github.com/your-username/don/pkg/httpclient"
)

type ExternalAPIAdapter struct {
	client *httpclient.Client
}

func NewExternalAPIAdapter(client *httpclient.Client) *ExternalAPIAdapter {
	return &ExternalAPIAdapter{client: client}
}

func (a *ExternalAPIAdapter) FetchData(ctx context.Context) (*httpclient.Response, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/data", nil)
	return a.client.Do(ctx, req)
}
```

## 🔍 Review and Suggested Improvements

### Flaws Identified
1. **Body Handling:** Original `http.Request` consumes `Body` on the first attempt. I added `GetBody` logic to ensure retries work for requests with payloads.
2. **Jitter Logic:** Simple `rand.Int63n` is used; for production, a more robust jitter algorithm might be desired, but this is a good start.
3. **Circular Dependency Risk:** Be careful with where the `Response` struct is defined. Since it's in `pkg/httpclient`, and use cases might need to inspect the response, use cases will depend on `pkg/httpclient`. This is acceptable for shared utility types.

### Suggestions
1. **Logging:** Integrate `pkg/logger` to log retry attempts and failures.
2. **Metrics:** Add support for Prometheus metrics (request count, duration, retry count).
3. **Custom Retry Policy:** Allow passing a custom function to determine if a request should be retried (e.g., based on error type).
4. **Transport Customization:** Add an option to provide a custom `http.RoundTripper` for features like connection pooling tuning or TLS config.
5. **Circuit Breaker:** Integrate a circuit breaker (e.g., `gobreaker`) to prevent cascading failures.

## 🚀 Next Steps
1. Approve this plan.
2. Implement `pkg/httpclient`.
3. Add unit tests for `pkg/httpclient` (mocking `http.RoundTripper`).
4. Implement ports and adapters.
