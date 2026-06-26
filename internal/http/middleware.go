package http

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	slogctx "github.com/veqryn/slog-context"
)

// useMiddlewares registers the middleware stack on e (outermost first).
func useMiddlewares(e *echo.Echo) {
	e.Use(middleware.Recover())                                                    // catch panics, return HTTP 500
	e.Use(requestID())                                                             // assign X-Request-Id + put it in the context
	e.Use(middleware.RequestLogger())                                              // structured access log via c.Logger()
	e.Use(middleware.Secure())                                                     // security response headers
	e.Use(middleware.CORS("*"))                                                    // permissive CORS (tighten origins in prod)
	e.Use(middleware.Gzip())                                                       // response compression
	e.Use(middleware.BodyLimit(maxRequestBody))                                    // reject oversized bodies
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rateLimit))) // throttle per client IP
	e.Use(middleware.ContextTimeout(requestTimeout))                               // bound handler execution time
}

// requestID assigns/propagates the X-Request-Id header and appends the id to the
// request context so slogctx.NewHandler can stamp it on every *Context log line.
func requestID() echo.MiddlewareFunc {
	return middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		RequestIDHandler: func(c *echo.Context, id string) {
			req := c.Request()
			c.SetRequest(req.WithContext(slogctx.Append(req.Context(), "request_id", id)))
		},
	})
}
