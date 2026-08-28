package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// serve starts the application HTTP server.
func (app *application) serve() error {
	// HTTP Server Configuration
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	// Goroutine for gracefully shutting down HTTP Server on SIGINT (Ctrl + C) and SIGTERM (pkill),
	// finishing any background tasks, and cancelling any ongoing worker routines.
	shutdownError := make(chan error)
	go func() {
		// Use a single buffered channel that blocks until a signal is received.
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit // Block here until signal is caught.
		app.logger.Info("shutting down server due to caught signal", "signal", s.String())

		// Allow HTTP server to close any remaining connections.
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		// BACKGROUND TASKS
		app.logger.Info("completing background tasks", "addr", srv.Addr)

		// Report worker goroutine
		if app.workerCancel != nil {
			app.workerCancel()
		}

		app.wg.Wait() // Block here until all background tasks are finished.
		shutdownError <- nil
	}()

	// Run HTTP Server
	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)
	err := srv.ListenAndServe()
	// NOTE: ErrServerClosed is expected when Shutdown is called in the preceding goroutine.
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Shutdown HTTP Server
	err = <-shutdownError // Block until shutdown signal is received.
	if err != nil {
		return err
	}
	app.logger.Info("stopped server", "addr", srv.Addr)

	return nil
}
