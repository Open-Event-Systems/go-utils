package gracefulshutdown

import (
	"os"
	"os/signal"
	"syscall"
)

// Handles graceful server shutdown when SIGINT/SIGTERM is received.
type GracefulShutdown struct {
	RunFunc  func() error
	Shutdown func() error
}

func (s *GracefulShutdown) Run() error {
	sigChan := make(chan os.Signal, 1)
	defer close(sigChan)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Reset(syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error, 1)

	go func() {
		defer close(errChan)
		errChan <- s.RunFunc()
	}()

	select {
	case <-sigChan:
		// received signal
		signal.Reset(syscall.SIGINT, syscall.SIGTERM)
		shutdownErr := s.Shutdown()
		<-errChan
		return shutdownErr
	case err := <-errChan:
		return err
	}
}
