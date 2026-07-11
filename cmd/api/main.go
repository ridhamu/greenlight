package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq" // To register the driver.
	"github.com/ridhamu/greenlight/internal/data"
	"github.com/ridhamu/greenlight/internal/mailer"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	mailer struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

type application struct {
	config config
	logger *slog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	var config config

	flag.IntVar(&config.port, "port", 4000, "API server port")
	flag.StringVar(&config.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&config.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&config.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&config.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&config.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max idle time")
	flag.Float64Var(&config.limiter.rps, "limiter-rps", 2, "Rate limiter maximum request per second")
	flag.IntVar(&config.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&config.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&config.mailer.host, "smtp-host", "sandbox.smtp.mailtrap.io", "smtp host")
	flag.IntVar(&config.mailer.port, "smtp-port", 2525, "smtp port")
	flag.StringVar(&config.mailer.username, "smtp-username", "eb119bfab9018e", "smtp username")
	flag.StringVar(&config.mailer.password, "smtp-password", "c97a8ca466e5f0", "smtp password")
	flag.StringVar(&config.mailer.sender, "smtp-sender", "Greenlight <no-reply@greenlight.alexedwards.net>", "smtp sender")

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
		models: data.NewModels(dbPool),
		mailer: mailer.New(config.mailer.host, config.mailer.port, config.mailer.username, config.mailer.password, config.mailer.sender),
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		logger.Info("i am here")
		os.Exit(1)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	dbPool, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	dbPool.SetMaxOpenConns(cfg.db.maxOpenConns)
	dbPool.SetMaxIdleConns(cfg.db.maxIdleConns)
	dbPool.SetConnMaxIdleTime(cfg.db.maxIdleTime)

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
