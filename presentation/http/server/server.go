// Package httpserver wires the HTTP router and the server lifecycle. It lives in
// the directory presentation/http/server.
package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// Server wraps an *http.Server with start/stop lifecycle helpers. Srv is
// exported so callers can configure it and tests can reach its handler. Lifecycle
// logging is the caller's responsibility (see cmd/main.go).
type Server struct {
	Srv *http.Server
}

// New builds a Server around a configured *http.Server.
func New(srv *http.Server) *Server {
	return &Server{Srv: srv}
}

// Start runs ListenAndServe in a goroutine, forwarding any non-graceful error to
// errChan (a clean shutdown via Stop is not reported as an error).
func (s *Server) Start(errChan chan<- error) {
	go func() {
		if err := s.Srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("failed to serve http: %w", err)
		}
	}()
}

// Stop gracefully shuts the server down, respecting ctx's deadline.
func (s *Server) Stop(ctx context.Context) error {
	if err := s.Srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shut down http server: %w", err)
	}

	return nil
}
