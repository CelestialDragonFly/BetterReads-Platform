package data

import (
	"context"
	"database/sql"
	"time"

	// Import the pq driver so that it can register itself with the database/sql
	// package. Note that we alias this import to the blank identifier, to stop the Go
	// compiler complaining that the package isn't being used.
	_ "github.com/lib/pq"
)

const driverName = "postgres"

type SQL struct {
	DB *sql.DB
}

type DBConfig struct {
	DataSourceName     string
	Timeout            time.Duration
	MaxOpenConnections int
	MaxIdleConnections int
	MaxIdleTime        time.Duration
}

// The New function returns a sql.DB connection pool.
func NewDatabase(config *DBConfig) (*SQL, error) {
	db, err := sql.Open(driverName, config.DataSourceName)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(config.MaxOpenConnections)
	db.SetMaxIdleConns(config.MaxIdleConnections)
	db.SetConnMaxIdleTime(config.MaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &SQL{DB: db}, nil
}
