package echoserver

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/dev-igorcarvalho/don/pkg/logger"
	"github.com/labstack/echo/v4"
)

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
