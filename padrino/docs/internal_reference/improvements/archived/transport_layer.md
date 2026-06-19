# HTTP Transport Layer Guidelines

This document outlines the standards and best practices for the HTTP transport layer in the **Don** project. These practices ensure security, reliability, and consistency across all API endpoints.

## 1. Strict JSON Decoding

### Problem
By default, Go's `json.Unmarshal` ignores unknown fields in the input. This can lead to:
- **Silenced Bugs**: Clients sending misspelled fields without being alerted.
- **Security Risks**: Potential mass-assignment vulnerabilities if internal fields are accidentally exposed in structs.
- **Resource Exhaustion**: Large request bodies can crash the service if not limited.

### Solution
Always use a decoder with strict settings and limit the request body size.

```go
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
    // Limit body size to 1MB
    r.Body = http.MaxBytesReader(w, r.Body, 1048576)

    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()

    if err := dec.Decode(dst); err != nil {
        return err
    }
    return nil
}
```

## 2. Double Validation

### Problem
Syntactic validation (e.g., checking if a field is present) is necessary but insufficient. Complex business rules (semantic validation) often depend on multiple fields or external state.

### Solution
Implement validation in two steps:
1. **Syntactic**: Use struct tags (e.g., `github.com/go-playground/validator`).
2. **Semantic**: Implement a `Validate()` method on the request struct for business logic.

```go
type CreateUserRequest struct {
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"required,gte=18"`
}

func (r *CreateUserRequest) Validate() error {
    if r.Email == "admin@blacklisted.com" {
        return errors.New("this email domain is not allowed")
    }
    return nil
}
```

## 3. Contextualization Middleware

### Problem
In distributed systems, tracing a single request across multiple services or even within a single service's logs is difficult without a unique identifier. Additionally, requests without timeouts can hang indefinitely, consuming resources.

### Solution
Use middleware to inject a Request ID and enforce a timeout.

```go
func ContextMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }
        
        // Propagate ID in headers and context
        w.Header().Set("X-Request-ID", requestID)
        ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
        
        // Apply timeout
        ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
        defer cancel()

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## 4. Header Management

### Problem
APIs are vulnerable to various attacks if headers are not strictly managed. Incorrect `Content-Type` can lead to parsing errors or security bypasses.

### Solution
Validate incoming headers and inject security headers in all responses.

**Mandatory Headers:**
- `Content-Type: application/json` (Validation)
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `Content-Security-Policy: default-src 'none'`

```go
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("Content-Type") != "application/json" && r.Method != http.MethodGet {
            http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
            return
        }

        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        next.ServeHTTP(w, r)
    })
}
```

## 5. Standardized Error Responses (RFC 7807)

### Problem
Inconsistent error formats make it hard for clients to handle failures programmatically.

### Solution
Follow [RFC 7807 (Problem Details for HTTP APIs)](https://tools.ietf.org/html/rfc7807). Map domain errors from `internal/core` to specific HTTP status codes.

```json
{
  "type": "https://example.com/probs/out-of-stock",
  "title": "Out of Stock",
  "status": 400,
  "detail": "Item XYZ is currently unavailable.",
  "instance": "/orders/123"
}
```

## 6. Rate Limiting (Opt-in)

### Problem
Endpoints can be abused by automated scripts, leading to service degradation.

### Solution
Implement an IP-based rate limiter using a "token bucket" algorithm and provide feedback via standard headers (`X-RateLimit-Limit`, `X-RateLimit-Remaining`).

```go
func RateLimitMiddleware(next http.Handler) http.Handler {
    // Simple implementation example using x/time/rate
    limiter := rate.NewLimiter(1, 5) // 1 req/s with burst of 5

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            w.Header().Set("Retry-After", "1")
            http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```
