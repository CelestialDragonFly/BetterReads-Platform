package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/celestialdragonfly/betterreads/internal/data"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrShelfNotFound   = errors.New("shelf not found")
	ErrBookNotFound    = errors.New("book not found in library")
	ErrShelfNameExists = errors.New("shelf name already exists")
	ErrInvalidShelfID  = errors.New("invalid shelf ID")
)

// Shelf operations

func (db *Client) CreateShelf(ctx context.Context, shelf *data.Shelf) (*data.Shelf, error) {
	query := `
		INSERT INTO shelves (id, name, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`

	err := db.DB.QueryRow(
		ctx,
		query,
		shelf.ID,
		shelf.Name,
		shelf.UserID,
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
	query := `
		UPDATE shelves
		SET name = $2, updated_at = $3
		WHERE id = $1
		RETURNING id, name, user_id, created_at, updated_at
	`

	var updatedShelf data.Shelf
	err := db.DB.QueryRow(
		ctx,
		query,
		shelf.ID,
		shelf.Name,
		shelf.UpdatedAt,
	).Scan(
		&updatedShelf.ID,
		&updatedShelf.Name,
		&updatedShelf.UserID,
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

func (db *Client) DeleteShelf(ctx context.Context, id string) error {
	query := `DELETE FROM shelves WHERE id = $1`
	result, err := db.DB.Exec(ctx, query, id)
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
		SELECT id, name, user_id, created_at, updated_at
		FROM shelves
		WHERE user_id = $1
		ORDER BY created_at ASC
	`

	rows, err := db.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("GetUserShelves: %w", err)
	}
	defer rows.Close()

	var shelves []*data.Shelf
	for rows.Next() {
		var s data.Shelf
		if err := rows.Scan(&s.ID, &s.Name, &s.UserID, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("GetUserShelves scan: %w", err)
		}
		shelves = append(shelves, &s)
	}
	return shelves, nil
}

func (db *Client) GetShelfBooks(ctx context.Context, shelfID string) ([]*data.LibraryBook, error) {
	// Join library_books with shelf_books
	// We might also want to fetch all shelf assignments for these books if we want to show them?
	// But for GetShelfBooks standard response typically just lists the books.
	// The LibraryBook struct has ShelfIDs. So we probably should populate it.
	// This makes it N+1 or a complex join.
	// Let's do a left join or aggregation.

	query := `
        SELECT lb.user_id, lb.book_id, lb.title, lb.author_name, lb.book_image, lb.rating, lb.source, lb.added_at, lb.updated_at,
               COALESCE(array_agg(sb2.shelf_id) FILTER (WHERE sb2.shelf_id IS NOT NULL), '{}') as shelf_ids
        FROM library_books lb
        JOIN shelf_books sb ON lb.book_id = sb.book_id AND lb.user_id = sb.user_id
        LEFT JOIN shelf_books sb2 ON lb.book_id = sb2.book_id AND lb.user_id = sb2.user_id
        WHERE sb.shelf_id = $1
        GROUP BY lb.user_id, lb.book_id, lb.title, lb.author_name, lb.book_image, lb.rating, lb.source, lb.added_at, lb.updated_at
    `

	books, err := db.queryLibraryBooks(ctx, query, shelfID)
	if err != nil {
		return nil, fmt.Errorf("GetShelfBooks: %w", err)
	}
	return books, nil
}

// Library Book operations

func (db *Client) UpdateLibraryBook(ctx context.Context, book *data.LibraryBook) error {
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
	_, err := db.DB.Exec(
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
	// This requires clearing existing shelf assignments and re-inserting, or smarter diffing.
	// For simplicity, let's delete all shelf assignments for this book and re-insert if shelf_ids provided.
	// Wait, UpdateLibraryBookRequest usually provides the full state or partial?
	// Proto says: "Add or update book in library".
	// If shelf_ids is empty, does it mean remove from all shelves?
	// Usually yes for a "Update" if it replaces the state.
	// But specific AddBookToShelf exists.
	// Let's assume UpdateLibraryBook handles the basic book details, and if shelf_ids is provided, it syncs them.

	if len(book.ShelfIDs) > 0 {
		// First, check if shelves exist and belong to user?
		// Assuming validation happens in service or we rely on FK constraints.

		// Remove existing associations
		// We only want to update shelves if we are sure that's the intent.
		// But if shelf_ids is passed, we should respect it.

		deleteQuery := `DELETE FROM shelf_books WHERE user_id = $1 AND book_id = $2`
		if _, err := db.DB.Exec(ctx, deleteQuery, book.UserID, book.BookID); err != nil {
			return fmt.Errorf("UpdateLibraryBook (clear shelves): %w", err)
		}

		insertQuery := `INSERT INTO shelf_books (shelf_id, user_id, book_id, added_at) VALUES ($1, $2, $3, $4)`
		for _, shelfID := range book.ShelfIDs {
			if _, err := db.DB.Exec(ctx, insertQuery, shelfID, book.UserID, book.BookID, time.Now()); err != nil {
				// Ignore duplicate key errors if any, but checking existence is better.
				// Actually relying on DB constraints is fine.
				return fmt.Errorf("UpdateLibraryBook (add shelf): %w", err)
			}
		}
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
	`
	books, err := db.queryLibraryBooks(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("GetUserLibrary: %w", err)
	}
	return books, nil
}

func (db *Client) AddBookToShelf(ctx context.Context, userID, bookID, shelfID string) error {
	// Check if book exists in library? FK constraint on shelf_books(user_id, book_id) -> library_books(user_id, book_id) handles this.
	// If book not in library, FK violation.
	// We should probably ensure the book is in library first or return error.

	query := `
		INSERT INTO shelf_books (shelf_id, user_id, book_id, added_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING
	`
	_, err := db.DB.Exec(ctx, query, shelfID, userID, bookID, time.Now())
	if err != nil {
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) {
			// Check for foreign key violation (book not in library)
			if pgErr.Code == ForeignKeyViolation {
				return ErrBookNotFound // or Shelf not found
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
