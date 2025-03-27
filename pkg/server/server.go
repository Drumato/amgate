package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/Drumato/amgate/pkg/config"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Server is the main struct for the server
type Server[T comparable] struct {
	e             *echo.Echo
	cfg           *config.Config
	logger        *slog.Logger
	CustomDepends *T
	K8sClient     client.Client
}

// Start starts the server
// this automatically starts the server and listens for interrupt signals
// to gracefully shut down the server.
func (s *Server[T]) Start(ctx context.Context) error {
	if s.logger == nil {
		s.logger = slog.Default()
	}
	port := lo.If(s.cfg.Server.Port != 0, s.cfg.Server.Port).Else(8080)
	host := lo.If(s.cfg.Server.Host != "", s.cfg.Server.Host).Else("") // all interfaces

	go func() {
		addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
		if err := s.e.Start(addr); err != nil && err != http.ErrServerClosed {
			s.logger.ErrorContext(ctx, "failed to start server", slog.String("error", err.Error()))
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.e.Shutdown(ctx); err != nil {
		s.e.Logger.Fatal(err)
	}

	return nil
}

// New creates a new server
func New[T comparable](e *echo.Echo, cfg *config.Config, options ...ServerOption[T]) *Server[T] {
	return &Server[T]{e: e, cfg: cfg}
}

// ServerOption is a function that modifies the server
type ServerOption[T comparable] func(*Server[T])

// WithCustomDepends sets the custom depends
func WithCustomDepends[T comparable](customDepends *T) ServerOption[T] {
	return func(s *Server[T]) {
		s.CustomDepends = customDepends
	}
}

// WithK8sClient sets the k8s client
func WithK8sClient[T comparable](k8sClient client.Client) ServerOption[T] {
	return func(s *Server[T]) {
		s.K8sClient = k8sClient
	}
}

func WithLogger[T comparable](logger *slog.Logger) ServerOption[T] {
	return func(s *Server[T]) {
		s.logger = logger
	}
}
