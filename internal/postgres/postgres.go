package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	ErrUnableToConnect = errors.New("unable to connect to Postgres DB")
	ErrUnableToPing    = errors.New("unable to ping Postgres DB")
)

type Client struct {
	DB *sql.DB
}

func NewClient(ctx context.Context, dsn string) (*Client, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("NewClient: %w", ErrUnableToConnect)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("NewClient: %w", ErrUnableToPing)
	}

	return &Client{DB: db}, nil
}
