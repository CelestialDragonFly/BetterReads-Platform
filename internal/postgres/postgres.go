package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/celestialdragonfly/betterreads/internal/data"
	"github.com/jackc/pgx/v5"
)

type API interface {
	ProfileCreate(ctx context.Context, profile *data.User) (*data.User, error)
	ProfileGet(ctx context.Context, id string) (*data.User, error)
	ProfileUpdate(ctx context.Context, id string, updates *data.User) (*data.User, error)
	ProfileDelete(ctx context.Context, id string) error
	GetUserByID(ctx context.Context, id string) (*data.User, error)
	FollowUser(ctx context.Context, followerID, followeeID string) error
	UnfollowUser(ctx context.Context, followerID, followeeID string) error

	// Library
	CreateShelf(ctx context.Context, shelf *data.Shelf) (*data.Shelf, error)
	UpdateShelf(ctx context.Context, shelf *data.Shelf) (*data.Shelf, error)
	DeleteShelf(ctx context.Context, userID, id string) error
	GetUserShelves(ctx context.Context, userID string) ([]*data.Shelf, error)
	GetShelfBooks(ctx context.Context, userID, shelfID string) ([]*data.LibraryBook, error)
	UpdateLibraryBook(ctx context.Context, book *data.LibraryBook) error
	RemoveLibraryBook(ctx context.Context, userID, bookID string) error
	GetUserLibrary(ctx context.Context, userID string) ([]*data.LibraryBook, error)
	AddBookToShelf(ctx context.Context, userID, bookID, shelfID string) error
	RemoveBookFromShelf(ctx context.Context, userID, bookID, shelfID string) error
}

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

var registers = []func(context.Context, *pgx.Conn) error{
	registerUser,
	registerFollows,
	registerLibrary,
}

func migrate(ctx context.Context, db *pgx.Conn) error {
	for _, f := range registers {
		err := f(ctx, db)
		if err != nil {
			return err
		}
	}
	return nil
}
