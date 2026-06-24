package httpmiddleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/reqctx"
)

const (
	requestIDHeader  = "X-Request-ID"
	requestIDByteLen = 16
)

// RequestID assigns a request id (honoring an inbound X-Request-ID), echoes it
// on the response, and stores it in the request context for downstream logging.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestIDHeader)
		if id == "" {
			id = newRequestID()
		}

		w.Header().Set(requestIDHeader, id)

		next.ServeHTTP(w, r.WithContext(reqctx.WithRequestID(r.Context(), id)))
	})
}

func newRequestID() string {
	b := make([]byte, requestIDByteLen)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}

	return hex.EncodeToString(b)
}
