package httputils

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// Run a [http.Server] with graceful shutdown handling.
func GracefulShutdown(ctx context.Context, logger *slog.Logger, server *http.Server) {
	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	var serverErr error
	doneChan := make(chan struct{})

	go func() {
		defer close(doneChan)
		serverErr = server.ListenAndServe()
	}()

	select {
	case <-sigCtx.Done():
		logger.InfoContext(sigCtx, "stopping server")
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
		<-doneChan
	case <-doneChan:
		stop()
	}

	if serverErr != nil && !errors.Is(serverErr, http.ErrServerClosed) {
		logger.ErrorContext(sigCtx, "server exited", slog.Any("err", serverErr))
	}
}
