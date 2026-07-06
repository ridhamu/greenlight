package main

import (
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

	// goroutine for checking signal SIGTERM and SIGINT
	go func() {
		quitChan := make(chan os.Signal, 1)

		signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)

		s := <-quitChan // this is blocking until we recheieved a signal

		app.logger.Info("caught signal", "signal", s.String())

		os.Exit(0)
	}()

	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)
	return srv.ListenAndServe()
}
