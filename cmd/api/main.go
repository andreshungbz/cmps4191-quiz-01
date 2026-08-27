package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/andreshungbz/cmps4191-quiz-01/internal/data"
	"github.com/andreshungbz/cmps4191-quiz-01/internal/vcs"
	_ "github.com/lib/pq"
)

var (
	version = vcs.Version()
)

// config stores the API server configuration.
type config struct {
	port               int           // API server port
	env                string        // (development|staging|production)
	reportDelay        time.Duration // SYNCHRONOUS API TESTING: field for report generation delay
	workerPollInterval time.Duration // interval for the report worker to poll for new jobs
	db                 struct {
		dsn          string        // data source name
		maxOpenConns int           // maximum number of open connections to the database
		maxIdleConns int           // maximum number of idle connections in the connection pool
		maxIdleTime  time.Duration // maximum amount of time a connection may be idle
	}
}

// application holds the dependencies for the HTTP handlers, helpers, middleware,
// etc. so that they are all accessible through dependency injection.
type application struct {
	config       config
	logger       *slog.Logger
	models       data.Models
	wg           sync.WaitGroup
	workerCancel context.CancelFunc
}

func main() {
	var cfg config

	// FLAGS

	// server flags
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.DurationVar(&cfg.reportDelay, "report-delay", 0, "Artificial report-generation delay")                            // SYNCHRONOUS API TESTING
	flag.DurationVar(&cfg.workerPollInterval, "worker-poll-interval", 250*time.Millisecond, "Worker queue-check interval") // ASYNCHRONOUS API TESTING

	// database flags
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	// version flag
	displayVersion := flag.Bool("version", false, "Display program version")

	flag.Parse()

	// display program version and exit if the version flag was passed
	if *displayVersion {
		fmt.Printf("version:\t%s\n", version)
		os.Exit(0)
	}

	// logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// DATABASE

	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connection pool established")

	// APPLICATION

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	workerCtx, cancelWorker := context.WithCancel(context.Background())
	app.workerCancel = cancelWorker
	defer cancelWorker()
	app.startReportWorker(workerCtx)

	// start the API server
	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

// openDB connects to the PostgreSQL database using the provided DSN and
// and returns a pointer to a handler to that database.
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// test the connection with a ping
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
