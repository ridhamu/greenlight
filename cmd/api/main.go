package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq" // To register the driver.
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
}

type application struct {
	config config
	logger *slog.Logger
}

func main() {
	var config config

	flag.IntVar(&config.port, "port", 4000, "API server port")
	flag.StringVar(&config.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&config.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	dbPool, err := openDB(config)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer func() {
		_ = dbPool.Close()
	}()
	logger.Info("database connection pool establish")

	app := application{
		config: config,
		logger: logger,
	}

	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("starting server", "addr", srv.Addr, "env", config.env)
	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}

func openDB(cfg config) (*sql.DB, error) {
	dbPool, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	contextWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = dbPool.PingContext(contextWithTimeout)
	if err != nil {
		_ = dbPool.Close()
		return nil, err
	}

	// return the connection pool
	return dbPool, nil
}
