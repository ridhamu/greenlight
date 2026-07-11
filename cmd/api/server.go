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

func (app *application) serve() error {
	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	// chanel for holding the error from calling method Shutdown(), if one exist
	shutDownErr := make(chan error)

	// goroutine for checking signal SIGTERM and SIGINT
	go func() {
		quitChan := make(chan os.Signal, 1)

		signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)

		s := <-quitChan // this is blocking until we recheieved a signal

		app.logger.Info("Shutting down the server", "signal", s.String())

		ctx, cf := context.WithTimeout(context.Background(), 30*time.Second)
		defer cf()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutDownErr <- err
		}

		app.logger.Info("completing background task", "addr", srv.Addr)

		app.wg.Wait()
		shutDownErr <- nil
	}()

	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// we wait for the actual response from chanel
	err = <-shutDownErr
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", "add", srv.Addr)
	return nil
}
