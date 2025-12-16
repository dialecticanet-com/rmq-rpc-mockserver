// Package components provides functionality for any components that run in the background and should stop gracefully.
package components

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Component interface defines the contract for a component.
// it is expected to be implemented by all background components that need to be started and stopped.
// Run is expected to block until the context is canceled or an error occurs.
type Component interface {
	Run(ctx context.Context) error
}

// Runner is a collection of components.
type Runner struct {
	termSig chan os.Signal
}

// NewRunner creates a new collection of components.
func NewRunner() (*Runner, error) {
	return &Runner{
		termSig: make(chan os.Signal, 1),
	}, nil
}

// Run starts all components concurrently and blocks until all components have exited.
func (c *Runner) Run(ctx context.Context, components ...Component) error {
	signal.Notify(c.termSig, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	wg := sync.WaitGroup{}
	wg.Add(len(components))
	chErr := make(chan error, len(components))

	cctx, cnl := context.WithCancel(ctx)
	for _, component := range components {
		go func() {
			defer wg.Done()
			if err := component.Run(cctx); err != nil {
				chErr <- err
			}
		}()
	}

	slog.InfoContext(cctx, "all components start initiated")
	allErrors := c.waitTermination(ctx, chErr)
	cnl()

	wg.Wait()
	close(chErr)
	slog.InfoContext(cctx, "all components stopped")

	for err := range chErr {
		allErrors = errors.Join(allErrors, err)
	}

	return allErrors
}

func (c *Runner) waitTermination(ctx context.Context, chErr <-chan error) error {
	for {
		select {
		case sig := <-c.termSig:
			slog.Info("os signal received, exiting...", "signal", sig.String())
			return nil
		case <-ctx.Done():
			slog.Info("global context canceled, exiting...")
			return nil
		case err := <-chErr:
			return err
		}
	}
}
