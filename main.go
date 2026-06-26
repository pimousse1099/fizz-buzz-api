package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type (
	fizzBuzzRequest struct {
		Str1  string `query:"str1" validate:"required"`
		Str2  string `query:"str2" validate:"required"`
		Int1  uint   `query:"int1" validate:"required"`
		Int2  uint   `query:"int2" validate:"required"`
		Limit uint   `query:"limit" validate:"required"`
	}
	fizzBuzzResponse []string
)

func (fr fizzBuzzRequest) String() string {
	return fmt.Sprintf("int1:%d_int2:%d_limit:%d_str1:%s_str2:%s", fr.Int1, fr.Int2, fr.Limit, fr.Str1, fr.Str2)
}

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
	e := getHTTPServer()

	// Cancel the start context on SIGINT/SIGTERM so StartConfig performs a graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	e.Logger.Info("starting server", "address", listenAddress)

	sc := echo.StartConfig{Address: listenAddress, GracefulTimeout: shutdownTimeout}

	err := sc.Start(ctx, e)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		e.Logger.Error("server terminated unexpectedly", "error", err)
		os.Exit(1)
	}

	e.Logger.Info("server stopped gracefully")
}

func getHTTPServer() *echo.Echo {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

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

	metricsCollector := &metricsCollector{}

	e.GET("/fizz-buzz", fizzBuzzHandler(metricsCollector)) // parameters passed as query string
	e.GET("/metrics", metricsHandler(metricsCollector))

	return e
}

// requestLogger bridges echo's RequestLogger middleware to the application slog logger.
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
			ctx := c.Request().Context()
			if v.Error != nil {
				logger.LogAttrs(ctx, slog.LevelError, "request failed", append(attrs, slog.String("error", v.Error.Error()))...)
				return nil
			}
			logger.LogAttrs(ctx, slog.LevelInfo, "request handled", attrs...)
			return nil
		},
	})
}

// =====================================================================================================================
// ============================================================ HANDLERS ===============================================
// =====================================================================================================================

func metricsHandler(mc *metricsCollector) echo.HandlerFunc {
	return func(c *echo.Context) error {
		if mc == nil || mc.RequestCounters == nil {
			return c.JSON(http.StatusOK, "no data collected yet")
		}
		sort.Sort(mc.RequestCounters)
		return c.JSON(http.StatusOK, mc.RequestCounters)
	}
}

func fizzBuzzHandler(mc *metricsCollector) echo.HandlerFunc {
	return func(c *echo.Context) error {
		// bind HTTP request (JSON) to fizzBuzzRequest
		req := new(fizzBuzzRequest)

		err := c.Bind(req)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		err = c.Validate(req)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		// increment request counter metric
		if mc != nil {
			mc.IncRequestCounter(req.String())
		}

		// call fizzbuzz controller and transform the response into an HTTP JSON response
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
