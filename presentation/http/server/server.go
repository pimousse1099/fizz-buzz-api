package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

// Server wraps an *http.Server with start/stop lifecycle helpers.
type Server struct {
	srv    *http.Server
	logger *slog.Logger
}

// New builds a Server around a configured *http.Server.
func New(srv *http.Server, logger *slog.Logger) *Server {
	return &Server{srv: srv, logger: logger}
}

// Start runs ListenAndServe in a goroutine, forwarding any non-graceful error
// to errChan.
func (s *Server) Start(errChan chan<- error) {
	go func() {
		s.logger.Info("starting HTTP server", "addr", s.srv.Addr)

		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("http server: %w", err)
		}
	}()
}

// Stop gracefully shuts the server down, respecting ctx's deadline.
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server")

	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	return nil
}
