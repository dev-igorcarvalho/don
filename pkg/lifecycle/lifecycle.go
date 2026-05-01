// ---
// title: Lifecycle Manager
// description: Coordinates graceful shutdown of application components using a LIFO approach.
// last_updated: 2026-05-01
// type: Implementation
// ---

// Package lifecycle provides a manager for coordinating the graceful shutdown
// of multiple components in the application.
package lifecycle

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dev-igorcarvalho/don/pkg/logger"
)

// Shutdownable defines the interface for components that require a graceful shutdown.
type Shutdownable interface {
	Shutdown(ctx context.Context) error
}

// Manager coordinates the graceful shutdown of multiple components.
type Manager struct {
	timeout    time.Duration
	components []component
	mu         sync.Mutex
}

type component struct {
	name string
	fn   Shutdownable
}

// NewManager creates a new lifecycle Manager.
func NewManager(timeout time.Duration) *Manager {
	return &Manager{
		timeout: timeout,
	}
}

// Register adds a component to the shutdown list.
// Components are shut down in the reverse order they were registered.
func (m *Manager) Register(name string, s Shutdownable) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.components = append(m.components, component{name: name, fn: s})
}

// ShutdownFunc is a function that implements the Shutdownable interface.
type ShutdownFunc func(ctx context.Context) error

// Shutdown calls the function.
func (f ShutdownFunc) Shutdown(ctx context.Context) error {
	return f(ctx)
}

// RegisterFunc adds a shutdown function to the manager.
func (m *Manager) RegisterFunc(name string, f func(ctx context.Context) error) {
	m.Register(name, ShutdownFunc(f))
}

// Wait blocks until a SIGINT or SIGTERM is received, then initiates shutdown.
func (m *Manager) Wait(ctx context.Context) error {
	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-sigCtx.Done()

	logger.Info(ctx, "shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	return m.shutdown(shutdownCtx)
}

func (m *Manager) shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error

	// Shutdown in reverse order (LIFO)
	for i := len(m.components) - 1; i >= 0; i-- {
		comp := m.components[i]
		logger.Info(ctx, "shutting down component", slog.String("name", comp.name))

		if err := comp.fn.Shutdown(ctx); err != nil {
			logger.Error(ctx, "failed to shutdown component", slog.String("name", comp.name), slog.Any("error", err))
			errs = append(errs, fmt.Errorf("%s: %w", comp.name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown completed with errors: %v", errs)
	}

	logger.Info(ctx, "graceful shutdown completed")
	return nil
}
