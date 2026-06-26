// Package main implements a simple fizz-buzz REST API server.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type (
	fizzBuzzRequest struct {
		Str1  string `json:"str1"  query:"str1"  validate:"required"`
		Str2  string `json:"str2"  query:"str2"  validate:"required"`
		Int1  uint   `json:"int1"  query:"int1"  validate:"required"`
		Int2  uint   `json:"int2"  query:"int2"  validate:"required"`
		Limit uint   `json:"limit" query:"limit" validate:"required"`
	}
	fizzBuzzResponse []string

	// statsResponse is the payload of the statistics endpoint: the parameters of
	// the most frequently requested fizz-buzz call and how many times it was made.
	statsResponse struct {
		RequestParams fizzBuzzRequest `json:"request_params"`
		Hits          uint            `json:"nb_hits"`
	}
)

type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}

const (
	listenAddress = ":8080"

	// maxRequestBody caps the accepted request body size.
	maxRequestBody = 1 << 20 // 1 MiB
	// rateLimit is the per-client request rate (requests/second).
	rateLimit = 20
	// requestTimeout bounds the time spent in the handler chain.
	requestTimeout = 10 * time.Second
	// shutdownTimeout bounds how long graceful shutdown waits for in-flight requests.
	shutdownTimeout = 10 * time.Second
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	e := getHTTPServer(logger)

	// Cancel the start context on SIGINT/SIGTERM so StartConfig performs a graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	logger.Info("starting server", "address", listenAddress)

	sc := echo.StartConfig{Address: listenAddress, GracefulTimeout: shutdownTimeout}

	err := sc.Start(ctx, e)

	// release the signal handler now that the server has returned
	stop()

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server terminated unexpectedly", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped gracefully")
}

func getHTTPServer(logger *slog.Logger) *echo.Echo {
	e := echo.New()
	e.Logger = logger // echo v5 logs through slog natively
	e.Validator = &customValidator{validator: validator.New()}

	// Middlewares (outermost first).
	e.Use(middleware.Recover())                                                    // catch panics, return HTTP 500
	e.Use(middleware.RequestID())                                                  // assign/propagate X-Request-ID
	e.Use(requestLogger(logger))                                                   // structured access log via slog
	e.Use(middleware.Secure())                                                     // security response headers
	e.Use(middleware.CORS("*"))                                                    // permissive CORS (tighten origins in prod)
	e.Use(middleware.Gzip())                                                       // response compression
	e.Use(middleware.BodyLimit(maxRequestBody))                                    // reject oversized bodies
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rateLimit))) // throttle per client IP
	e.Use(middleware.ContextTimeout(requestTimeout))                               // bound handler execution time

	stats := newMetricsCollector()

	e.GET("/fizz-buzz", fizzBuzzHandler(stats)) // parameters passed as query string
	e.GET("/metrics", metricsHandler(stats))

	return e
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

// =====================================================================================================================
// ============================================================ HANDLERS ===============================================
// =====================================================================================================================

func metricsHandler(stats *metricsCollector) echo.HandlerFunc {
	return func(c *echo.Context) error {
		req, hits, ok := stats.top()
		if !ok {
			return c.JSON(http.StatusOK, "no data collected yet")
		}

		return c.JSON(http.StatusOK, statsResponse{RequestParams: req, Hits: hits})
	}
}

func fizzBuzzHandler(stats *metricsCollector) echo.HandlerFunc {
	return func(c *echo.Context) error {
		// bind query parameters into fizzBuzzRequest
		req := new(fizzBuzzRequest)

		err := c.Bind(req)
		if err != nil {
			// return the error so it flows through the error handler and is logged
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		err = c.Validate(req)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		stats.record(*req)

		return c.JSON(http.StatusOK, fizzBuzzController(req))
	}
}

func fizzBuzzController(fizzBuzzReq *fizzBuzzRequest) *fizzBuzzResponse {
	fizzBuzzResp := make(fizzBuzzResponse, fizzBuzzReq.Limit)

	for i := uint(1); i <= fizzBuzzReq.Limit; i++ {
		res := ""

		if i%fizzBuzzReq.Int1 == 0 {
			res += fizzBuzzReq.Str1
		}

		if i%fizzBuzzReq.Int2 == 0 {
			res += fizzBuzzReq.Str2
		}

		if res != "" {
			fizzBuzzResp[i-1] = res

			continue
		}

		fizzBuzzResp[i-1] = strconv.FormatUint(uint64(i), 10)
	}

	return &fizzBuzzResp
}
