package http

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// useMiddlewares registers the middleware stack on e (outermost first).
func useMiddlewares(e *echo.Echo) {
	e.Use(middleware.Recover())                                                    // catch panics, return HTTP 500
	e.Use(middleware.RequestID())                                                  // assign/propagate X-Request-Id
	e.Use(middleware.RequestLogger())                                              // structured access log via c.Logger()
	e.Use(middleware.Secure())                                                     // security response headers
	e.Use(middleware.CORS("*"))                                                    // permissive CORS (tighten origins in prod)
	e.Use(middleware.Gzip())                                                       // response compression
	e.Use(middleware.BodyLimit(maxRequestBody))                                    // reject oversized bodies
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rateLimit))) // throttle per client IP
	e.Use(middleware.ContextTimeout(requestTimeout))                               // bound handler execution time
}
