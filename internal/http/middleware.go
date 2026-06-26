package http

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// useMiddlewares registers the middleware stack on e (outermost first).
func useMiddlewares(e *echo.Echo, logger *slog.Logger) {
	e.Use(middleware.Recover())                                                    // catch panics, return HTTP 500
	e.Use(middleware.RequestID())                                                  // assign/propagate X-Request-ID
	e.Use(requestLogger(logger))                                                   // structured access log via slog
	e.Use(middleware.Secure())                                                     // security response headers
	e.Use(middleware.CORS("*"))                                                    // permissive CORS (tighten origins in prod)
	e.Use(middleware.Gzip())                                                       // response compression
	e.Use(middleware.BodyLimit(maxRequestBody))                                    // reject oversized bodies
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rateLimit))) // throttle per client IP
	e.Use(middleware.ContextTimeout(requestTimeout))                               // bound handler execution time
}

// requestLogger bridges echo's RequestLogger middleware to the application slog
// logger. It logs one structured line per request and escalates the level for
// failures (4xx -> warn, 5xx -> error), attaching the error when present.
func requestLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogMethod:    true,
		LogURI:       true,
		LogStatus:    true,
		LogLatency:   true,
		LogRequestID: true,
		HandleError:  true, // let the global error handler set the status before logging
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			attrs := []slog.Attr{
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
				slog.Duration("latency", v.Latency),
				slog.String("request_id", v.RequestID),
			}

			level := slog.LevelInfo
			msg := "request handled"

			if v.Error != nil {
				attrs = append(attrs, slog.String("error", v.Error.Error()))
				msg = "request failed"
				level = slog.LevelWarn

				if v.Status >= http.StatusInternalServerError {
					level = slog.LevelError
				}
			}

			logger.LogAttrs(c.Request().Context(), level, msg, attrs...)

			return nil
		},
	})
}
