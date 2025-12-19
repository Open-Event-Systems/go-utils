package gracefulshutdown

import (
	"context"
	"net/http"
	"time"
)

const defaultTimeout = 5 * time.Second

// Handles graceful shutdown of a http.Server when SIGINT/SIGTERM is received.
type GracefulHTTPShutdown struct {
	Server          *http.Server
	Context         context.Context
	ShutdownTimeout time.Duration
}

func (s *GracefulHTTPShutdown) ListenAndServe() error {
	return s.run(s.Server.ListenAndServe)
}

func (s *GracefulHTTPShutdown) ListenAndServeTLS(certFile string, keyFile string) error {
	return s.run(func() error {
		return s.Server.ListenAndServeTLS(certFile, keyFile)
	})
}

func (s *GracefulHTTPShutdown) run(runFunc func() error) error {
	defer s.Server.Close()

	gs := GracefulShutdown{
		RunFunc: runFunc,
		Shutdown: func() error {
			timeout := s.ShutdownTimeout
			if timeout == 0 {
				timeout = defaultTimeout
			}

			shutdownCtx, cancel := context.WithTimeout(s.Context, timeout)
			defer cancel()

			return s.Server.Shutdown(shutdownCtx)
		},
	}

	return gs.Run()
}
