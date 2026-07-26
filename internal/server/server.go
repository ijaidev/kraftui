package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	embedfs "github.com/ijaidev/kraftui/internal/ui"
	"github.com/ijaidev/kraftui/log"
)

type Server struct {
	version string
	port    int
}

func (s *Server) Listen(ctx context.Context) error {
	ui, err := embedfs.UiHandler()
	if err != nil {
		return fmt.Errorf("embed frontend: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"status":"ok","version":%q}`, s.version)
	})
	mux.Handle("/", ui)

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	listner, err := s.getListner()
	if err != nil {
		return err
	}
	errCh := make(chan error, 1)
	go func() {
		log.G.Info("server listening", "version", s.version, "url", fmt.Sprintf("http://localhost:%d", s.port))
		if err := srv.Serve(listner); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		log.G.Info("received interruption, shutting down")
	case err := <-errCh:
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	err = srv.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("error shutting down the server: %w", err)
	}
	return nil
}

func (s *Server) getListner() (net.Listener, error) {
	maxAttempts := 5

	if s.port != 0 {
		address := fmt.Sprintf(":%d", s.port)
		return net.Listen("tcp", address)
	}

	startPort := 5200
	for i := range maxAttempts {
		currentPort := startPort + i
		address := fmt.Sprintf(":%d", currentPort)

		listener, err := net.Listen("tcp", address)
		if err == nil {
			s.port = currentPort
			return listener, nil
		}

		// Log only if the failure reason is actually "port occupied"
		if isPortOccupiedError(err) {
			log.G.Warn("port is occupied, checking next", "port", currentPort)
		} else {
			// Fail immediately if it's a critical error (like permission denied)
			return nil, fmt.Errorf("critical binding error on port %d: %w", currentPort, err)
		}
	}

	return nil, fmt.Errorf("exhausted all %d ports starting from %d", maxAttempts, startPort)
}

func NewServer(port int, version string) *Server {
	return &Server{
		port:    port,
		version: version,
	}
}
