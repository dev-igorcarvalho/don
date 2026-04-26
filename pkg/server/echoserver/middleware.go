package echoserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/dev-igorcarvalho/don/pkg/logger"
	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

type rateLimitedClient struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimitMiddleware returns a middleware that limits the number of requests
// per IP address using the token bucket algorithm. It includes a background
// worker that cleans up idle entries to prevent memory leaks, using the provided
// context for graceful shutdown.
func RateLimitMiddleware(ctx context.Context, rps float64, burst int) echo.MiddlewareFunc {
	var limiters sync.Map

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				limiters.Range(func(key, value any) bool {
					c := value.(*rateLimitedClient)
					if time.Since(c.lastSeen) > 1*time.Hour {
						limiters.Delete(key)
					}
					return true
				})
			}
		}
	}()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			now := time.Now()

			l, _ := limiters.LoadOrStore(ip, &rateLimitedClient{
				limiter:  rate.NewLimiter(rate.Limit(rps), burst),
				lastSeen: now,
			})

			client := l.(*rateLimitedClient)
			client.lastSeen = now

			if !client.limiter.Allow() {
				c.Response().Header().Set("Retry-After", "1")
				return echo.NewHTTPError(http.StatusTooManyRequests, "Too Many Requests")
			}

			return next(c)
		}
	}
}

// SecurityHeadersMiddleware returns a middleware that validates incoming headers
// and injects security headers in all responses.
// If no allowedContentTypes are provided, it defaults to application/json.
func SecurityHeadersMiddleware(allowedContentTypes ...string) echo.MiddlewareFunc {
	if len(allowedContentTypes) == 0 {
		allowedContentTypes = []string{echo.MIMEApplicationJSON}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			// Validate Content-Type for non-GET requests
			if req.Method != http.MethodGet {
				contentType := req.Header.Get(echo.HeaderContentType)
				allowed := false
				for _, t := range allowedContentTypes {
					if contentType == t {
						allowed = true
						break
					}
				}

				if !allowed {
					return echo.NewHTTPError(http.StatusUnsupportedMediaType, "Unsupported Media Type")
				}
			}

			res := c.Response()

			res.Header().Set("X-Content-Type-Options", "nosniff")
			res.Header().Set("X-Frame-Options", "DENY")
			res.Header().Set("Content-Security-Policy", "default-src 'none'")

			return next(c)
		}
	}
}

// ContextFromHeaderMiddleware returns a middleware that extracts the specified headers
// from the request and injects them into the request context.
func ContextFromHeaderMiddleware(headers ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx := req.Context()

			for _, h := range headers {
				if val := req.Header.Get(h); val != "" {
					ctx = context.WithValue(ctx, h, val)
				}
			}

			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	}
}

// LoggerMiddleware returns a middleware that logs HTTP requests with additional base attributes.
func LoggerMiddleware(baseAttrs ...slog.Attr) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			stop := time.Now()

			req := c.Request()
			res := c.Response()

			// Build dynamic attributes for this specific request
			attrs := append([]slog.Attr{}, baseAttrs...)
			attrs = append(attrs,
				slog.String("method", req.Method),
				slog.String("path", req.URL.Path),
				slog.Int("status", res.Status),
				slog.Duration("latency", stop.Sub(start)),
				slog.String("ip", c.RealIP()),
				slog.String("user_agent", req.UserAgent()),
			)

			if err != nil {
				attrs = append(attrs, slog.String("error", err.Error()))
				slog.LogAttrs(req.Context(), slog.LevelError, "request failed", attrs...)
			} else {
				slog.LogAttrs(req.Context(), slog.LevelInfo, "request processed", attrs...)
			}

			return nil
		}
	}
}

// RecoveryMiddleware returns a middleware that recovers from panics and logs them.
func RecoveryMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}

					stack := make([]byte, 4096)
					length := runtime.Stack(stack, false)

					attrs := []slog.Attr{
						slog.String("error", err.Error()),
						slog.String("stack", string(stack[:length])),
					}

					logger.Error(c.Request().Context(), "panic recovered", attrs...)

					c.Error(err)
				}
			}()
			return next(c)
		}
	}
}
