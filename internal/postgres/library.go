package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/celestialdragonfly/betterreads/internal/data"
	"github.com/celestialdragonfly/betterreads/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrShelfNotFound            = errors.New("shelf not found")
	ErrBookNotFound             = errors.New("book not found in library")
	ErrShelfNameExists          = errors.New("shelf name already exists")
	ErrInvalidShelfID           = errors.New("invalid shelf ID")
	ErrCannotDeleteDefaultShelf = errors.New("cannot delete default shelf")
	ErrCannotUpdateDefaultShelf = errors.New("cannot update default shelf")
)

// Shelf operations

func (db *Client) CreateShelf(ctx context.Context, shelf *data.Shelf) (*data.Shelf, error) {
	query := `
		INSERT INTO shelves (id, name, user_id, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`

	err := db.DB.QueryRow(
		ctx,
		query,
		shelf.ID,
		shelf.Name,
		shelf.UserID,
		shelf.IsDefault,
		shelf.CreatedAt,
		shelf.UpdatedAt,
	).Scan(&shelf.CreatedAt, &shelf.UpdatedAt)
	if err != nil {
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) {
			if pgErr.Code == UniqueViolation {
				return nil, ErrShelfNameExists
			}
		}
		return nil, fmt.Errorf("CreateShelf: %w", err)
	}

	return shelf, nil
}

func (db *Client) UpdateShelf(ctx context.Context, shelf *data.Shelf) (*data.Shelf, error) {
	// First check if this is a default shelf
	checkQuery := `SELECT is_default FROM shelves WHERE id = $1 AND user_id = $2`
	var isDefault bool
	err := db.DB.QueryRow(ctx, checkQuery, shelf.ID, shelf.UserID).Scan(&isDefault)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrShelfNotFound
		}
		return nil, fmt.Errorf("UpdateShelf check: %w", err)
	}

	if isDefault {
		return nil, ErrCannotUpdateDefaultShelf
	}

	query := `
		UPDATE shelves
		SET name = $2, updated_at = $3
		WHERE id = $1 AND user_id = $4
		RETURNING id, name, user_id, is_default, created_at, updated_at
	`

	var updatedShelf data.Shelf
	err = db.DB.QueryRow(
		ctx,
		query,
		shelf.ID,
		shelf.Name,
		shelf.UpdatedAt,
		shelf.UserID,
	).Scan(
		&updatedShelf.ID,
		&updatedShelf.Name,
		&updatedShelf.UserID,
		&updatedShelf.IsDefault,
		&updatedShelf.CreatedAt,
		&updatedShelf.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrShelfNotFound
		}
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) {
			if pgErr.Code == UniqueViolation {
				return nil, ErrShelfNameExists
			}
		}
		return nil, fmt.Errorf("UpdateShelf: %w", err)
	}

	return &updatedShelf, nil
}

func (db *Client) DeleteShelf(ctx context.Context, userID, id string) error {
	// First check if this is a default shelf
	checkQuery := `SELECT is_default FROM shelves WHERE id = $1 AND user_id = $2`
	var isDefault bool
	err := db.DB.QueryRow(ctx, checkQuery, id, userID).Scan(&isDefault)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrShelfNotFound
		}
		return fmt.Errorf("DeleteShelf check: %w", err)
	}

	if isDefault {
		return ErrCannotDeleteDefaultShelf
	}

	query := `DELETE FROM shelves WHERE id = $1 AND user_id = $2`
	result, err := db.DB.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("DeleteShelf: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrShelfNotFound
	}
	return nil
}

func (db *Client) GetUserShelves(ctx context.Context, userID string) ([]*data.Shelf, error) {
	query := `
		SELECT id, name, user_id, is_default, created_at, updated_at
		FROM shelves
		WHERE user_id = $1
		ORDER BY is_default DESC, created_at ASC
	`

	rows, err := db.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("GetUserShelves: %w", err)
	}
	defer rows.Close()

	var shelves []*data.Shelf
	for rows.Next() {
		var s data.Shelf
		if err := rows.Scan(&s.ID, &s.Name, &s.UserID, &s.IsDefault, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("GetUserShelves scan: %w", err)
		}
		shelves = append(shelves, &s)
	}
	return shelves, nil
}

func (db *Client) GetShelfBooks(ctx context.Context, userID, shelfID string) ([]*data.LibraryBook, error) {
	// Use CTE to avoid double-joining shelf_books table
	query := `
		WITH shelf_book_ids AS (
			SELECT book_id
			FROM shelf_books
			WHERE shelf_id = $1 AND user_id = $2
		)
		SELECT lb.user_id, lb.book_id, lb.title, lb.author_name, lb.book_image, lb.rating, lb.source, lb.added_at, lb.updated_at,
			   COALESCE(array_agg(sb.shelf_id) FILTER (WHERE sb.shelf_id IS NOT NULL), '{}') as shelf_ids
		FROM library_books lb
		INNER JOIN shelf_book_ids sbi ON lb.book_id = sbi.book_id
		LEFT JOIN shelf_books sb ON lb.book_id = sb.book_id AND lb.user_id = sb.user_id
		WHERE lb.user_id = $2
		GROUP BY lb.user_id, lb.book_id, lb.title, lb.author_name, lb.book_image, lb.rating, lb.source, lb.added_at, lb.updated_at
		ORDER BY lb.added_at DESC
	`

	books, err := db.queryLibraryBooks(ctx, query, shelfID, userID)
	if err != nil {
		return nil, fmt.Errorf("GetShelfBooks: %w", err)
	}
	return books, nil
}

// Library Book operations

func (db *Client) UpdateLibraryBook(ctx context.Context, book *data.LibraryBook) error {
	tx, err := db.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer func(ctx context.Context, tx pgx.Tx) {
		if rollBackErr := tx.Rollback(ctx); rollBackErr != nil {
			logger.Error("rolling back transaction. user_id: %s, book_id: %s, err: %w",
				book.UserID, book.BookID, rollBackErr)
		}
	}(ctx, tx)

	// Upsert book
	query := `
		INSERT INTO library_books (user_id, book_id, title, author_name, book_image, rating, source, added_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id, book_id) DO UPDATE
		SET title = EXCLUDED.title,
			author_name = EXCLUDED.author_name,
			book_image = EXCLUDED.book_image,
			rating = EXCLUDED.rating,
			source = EXCLUDED.source,
			updated_at = EXCLUDED.updated_at
	`
	_, err = tx.Exec(
		ctx,
		query,
		book.UserID,
		book.BookID,
		book.Title,
		book.AuthorName,
		book.BookImage,
		book.Rating,
		book.Source,
		book.AddedAt,
		book.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("UpdateLibraryBook: %w", err)
	}

	// Handle shelf assignments
	// We always clear and re-insert to ensure state matches the request.
	deleteQuery := `DELETE FROM shelf_books WHERE user_id = $1 AND book_id = $2`
	if _, err := tx.Exec(ctx, deleteQuery, book.UserID, book.BookID); err != nil {
		return fmt.Errorf("UpdateLibraryBook (clear shelves): %w", err)
	}

	if len(book.ShelfIDs) > 0 {
		insertQuery := `INSERT INTO shelf_books (shelf_id, user_id, book_id, added_at) VALUES ($1, $2, $3, $4)`
		for _, shelfID := range book.ShelfIDs {
			if _, err := tx.Exec(ctx, insertQuery, shelfID, book.UserID, book.BookID, time.Now()); err != nil {
				// We return error on invalid shelf_id (foreign key violation)
				return fmt.Errorf("UpdateLibraryBook (add shelf): %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

func (db *Client) RemoveLibraryBook(ctx context.Context, userID, bookID string) error {
	query := `DELETE FROM library_books WHERE user_id = $1 AND book_id = $2`
	result, err := db.DB.Exec(ctx, query, userID, bookID)
	if err != nil {
		return fmt.Errorf("RemoveLibraryBook: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrBookNotFound
	}
	return nil
}

func (db *Client) GetUserLibrary(ctx context.Context, userID string) ([]*data.LibraryBook, error) {
	query := `
		SELECT lb.user_id, lb.book_id, lb.title, lb.author_name, lb.book_image, lb.rating, lb.source, lb.added_at, lb.updated_at,
			   COALESCE(array_agg(sb.shelf_id) FILTER (WHERE sb.shelf_id IS NOT NULL), '{}') as shelf_ids
		FROM library_books lb
		LEFT JOIN shelf_books sb ON lb.book_id = sb.book_id AND lb.user_id = sb.user_id
		WHERE lb.user_id = $1
		GROUP BY lb.user_id, lb.book_id
		ORDER BY lb.added_at DESC
	`
	books, err := db.queryLibraryBooks(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("GetUserLibrary: %w", err)
	}
	return books, nil
}

func (db *Client) AddBookToShelf(ctx context.Context, userID, bookID, shelfID string) error {
	query := `
		INSERT INTO shelf_books (shelf_id, user_id, book_id, added_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING
	`
	_, err := db.DB.Exec(ctx, query, shelfID, userID, bookID, time.Now())
	if err != nil {
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) {
			if pgErr.Code == ForeignKeyViolation {
				// Check which constraint was violated
				switch pgErr.ConstraintName {
				case "shelf_books_user_id_book_id_fkey":
					return ErrBookNotFound
				case "shelf_books_shelf_id_fkey":
					return ErrShelfNotFound
				default:
					return fmt.Errorf("AddBookToShelf FK violation: %w", err)
				}
			}
		}
		return fmt.Errorf("AddBookToShelf: %w", err)
	}
	return nil
}

func (db *Client) RemoveBookFromShelf(ctx context.Context, userID, bookID, shelfID string) error {
	query := `DELETE FROM shelf_books WHERE user_id = $1 AND book_id = $2 AND shelf_id = $3`
	_, err := db.DB.Exec(ctx, query, userID, bookID, shelfID)
	if err != nil {
		return fmt.Errorf("RemoveBookFromShelf: %w", err)
	}
	return nil
}

// Table registration

func (db *Client) queryLibraryBooks(ctx context.Context, query string, args ...any) ([]*data.LibraryBook, error) {
	rows, err := db.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []*data.LibraryBook
	for rows.Next() {
		var b data.LibraryBook
		if err := rows.Scan(
			&b.UserID,
			&b.BookID,
			&b.Title,
			&b.AuthorName,
			&b.BookImage,
			&b.Rating,
			&b.Source,
			&b.AddedAt,
			&b.UpdatedAt,
			&b.ShelfIDs,
		); err != nil {
			return nil, fmt.Errorf("scan library book: %w", err)
		}
		books = append(books, &b)
	}
	return books, nil
}

func registerLibrary(ctx context.Context, db *pgx.Conn) error {
	createShelvesTable := `
	CREATE TABLE IF NOT EXISTS shelves (
		id UUID PRIMARY KEY,
		name TEXT NOT NULL,
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		is_default BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE(user_id, name)
	);
	`
	if _, err := db.Exec(ctx, createShelvesTable); err != nil {
		return fmt.Errorf("failed to create shelves table: %w", err)
	}

	createLibraryBooksTable := `
	CREATE TABLE IF NOT EXISTS library_books (
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		book_id TEXT NOT NULL,
		title TEXT NOT NULL,
		author_name TEXT NOT NULL,
		book_image TEXT,
		rating INTEGER,
		source INTEGER,
		added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		PRIMARY KEY (user_id, book_id)
	);
	`
	if _, err := db.Exec(ctx, createLibraryBooksTable); err != nil {
		return fmt.Errorf("failed to create library_books table: %w", err)
	}

	createShelfBooksTable := `
	CREATE TABLE IF NOT EXISTS shelf_books (
		shelf_id UUID NOT NULL REFERENCES shelves(id) ON DELETE CASCADE,
		user_id UUID NOT NULL,
		book_id TEXT NOT NULL,
		added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		PRIMARY KEY (shelf_id, book_id),
		FOREIGN KEY (user_id, book_id) REFERENCES library_books(user_id, book_id) ON DELETE CASCADE
	);
	`
	if _, err := db.Exec(ctx, createShelfBooksTable); err != nil {
		return fmt.Errorf("failed to create shelf_books table: %w", err)
	}

	return nil
}
