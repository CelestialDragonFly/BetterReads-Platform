package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var (
	ErrUnableToConnect = errors.New("unable to connect to Postgres DB")
	ErrUnableToPing    = errors.New("unable to ping Postgres DB")
)

type Client struct {
	DB *pgx.Conn
}

func NewClient(ctx context.Context, dsn string) (*Client, error) {
	db, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("NewClient postgress: %w", ErrUnableToConnect)
	}

	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("NewClient postgress: %w", ErrUnableToPing)
	}

	if err := migrate(ctx, db); err != nil {
		return nil, fmt.Errorf("NewClient postgress: %w", err)

	}
	return &Client{DB: db}, nil
}

var (
	registers = []func(context.Context, *pgx.Conn) error{
		registerUser,
	}
)

func migrate(ctx context.Context, db *pgx.Conn) error {
	for _, f := range registers {
		err := f(ctx, db)
		if err != nil {
			return err
		}
	}
	return nil
}
