package httpmiddleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/reqctx"
)

// statusRecorder captures the response status code for logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Logging logs one structured line per request with method, path, status and duration.
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(sr, r)

			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", sr.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", reqctx.RequestID(r.Context()),
			)
		})
	}
}
