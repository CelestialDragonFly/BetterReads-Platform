package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/celestialdragonfly/betterreads-platform/internal/data"
	"github.com/celestialdragonfly/betterreads-platform/internal/dependency/google"
	"github.com/celestialdragonfly/betterreads-platform/internal/package/env"
	"github.com/celestialdragonfly/betterreads-platform/internal/package/log"
	"github.com/celestialdragonfly/betterreads-platform/internal/server"
)

const version = "0.0.1"

func run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	var (
		port        = env.Int("HTTP_PORT", 4000, "API server port")
		environment = env.String("DEPLOYMENT_ENV", "development", "Environment (development|staging|production)")

		// Google API Envs
		googleBookAPIKey = env.String("GOOGLE_BOOK_API_KEY", "", "API Key for the Google Books dependency.")

		// Database Envs
		databaseDSN                = env.String("DATABASE_DSN", "", "PostgreSQL DSN")
		databaseTimeout            = env.Duration("DATABASE_TIMEOUT", 5*time.Second, "Create a context with a 5-second timeout deadline.")
		databaseMaxOpenConnections = env.Int("DATABASE_MAX_OPEN_CONNECTIONS", 25, "PostgreSQL max open connections")
		databaseMaxIdleConnections = env.Int("DATABASE_MAX_IDLE_CONNECTIONS", 25, "PostgreSQL max idle connections")
		databaseMaxIdleTime        = env.Duration("DATABASE_MAX_IDLE_CONNECTIONS", 15*time.Minute, "PostgreSQL max connection idle time")

		cfg = server.Config{
			Port:    port,
			Env:     environment,
			Version: version,
		}
	)

	database, err := data.NewDatabase(&data.DBConfig{
		DataSourceName:     databaseDSN,
		Timeout:            databaseTimeout,
		MaxOpenConnections: databaseMaxOpenConnections,
		MaxIdleConnections: databaseMaxIdleConnections,
		MaxIdleTime:        databaseMaxIdleTime,
	})
	if err != nil {
		panic(fmt.Errorf("error connecting to database %w", err))
	}
	googleBooksAPI := google.NewAPI(googleBookAPIKey)

	logger := log.NewLogger(w)
	app := server.NewBetterReads(logger, database, googleBooksAPI, &cfg)
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      app.Handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     logger.NewErrorLogger(),
	}

	go func() {
		logger.Info(fmt.Sprintf("listening on %s", srv.Addr), nil)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("error listening and serving", log.Fields{"error": err})
		}
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()
	return nil

}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
