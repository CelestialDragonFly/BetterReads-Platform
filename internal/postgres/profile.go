package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/celestialdragonfly/betterreads/internal/data"
	"github.com/celestialdragonfly/betterreads/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrUserNameExists          = errors.New("ProfileCreate: username already exists")
	ErrEmailExists             = errors.New("ProfileCreate: email already exists")
	ErrUnknownUniqueConstraint = errors.New("ProfileCreate: unique constraint violation")
	ErrInsertUser              = errors.New("ProfileCreate: failed to insert user")
	ErrUserNotFound            = errors.New("ProfileGet: user not found")
	ErrGetUser                 = errors.New("ProfileGet: failed to retrieve user")
	ErrUpdateUser              = errors.New("ProfileUpdate: failed to update user")
	ErrDeleteUser              = errors.New("ProfileDelete: failed to delete user")
)

func (db *Client) ProfileCreate(ctx context.Context, profile *data.User) (*data.User, error) {
	// Start a transaction to ensure user and default shelves are created atomically
	tx, err := db.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("ProfileCreate: failed to start transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			logger.Error("ProfileCreate: failed to rollback transaction: %w", rollbackErr)
		}
	}()

	query := `
		INSERT INTO users (id, username, first_name, last_name, email, profile_photo)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at;
	`

	var createdAt time.Time

	err = tx.QueryRow(
		ctx,
		query,
		profile.ID,
		profile.Username,
		profile.FirstName,
		profile.LastName,
		profile.Email,
		profile.ProfilePhotoURL,
	).Scan(&createdAt)
	if err != nil {
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) {
			if pgErr.Code == UniqueViolation {
				switch pgErr.ConstraintName {
				case "users_username_key":
					return nil, ErrUserNameExists
				case "users_email_key":
					return nil, ErrEmailExists
				default:
					return nil, fmt.Errorf("%w: constaint_name: %s", ErrUnknownUniqueConstraint, pgErr.ConstraintName)
				}
			}
			return nil, ErrUnknownUniqueConstraint
		}
		return nil, fmt.Errorf("%w: %w", ErrInsertUser, err)
	}

	profile.CreatedAt = createdAt

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("ProfileCreate: failed to commit transaction: %w", err)
	}

	return profile, nil
}

func (db *Client) ProfileGet(ctx context.Context, id string) (*data.User, error) {
	query := `
		SELECT id, username, first_name, last_name, email, profile_photo, created_at
		FROM users
		WHERE id = $1;
	`

	var user data.User

	err := db.DB.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.ProfilePhotoURL,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: %w", ErrGetUser, err)
	}

	return &user, nil
}

func (db *Client) ProfileUpdate(ctx context.Context, id string, updates *data.User) (*data.User, error) {
	// Build dynamic SET clause
	setClauses := []string{}
	args := []any{}
	argPos := 1

	if updates.GetUsername() == "" {
		setClauses = append(setClauses, fmt.Sprintf("username = $%d", argPos))
		args = append(args, updates.Username)
		argPos++
	}
	if updates.GetFirstName() == "" {
		setClauses = append(setClauses, fmt.Sprintf("first_name = $%d", argPos))
		args = append(args, updates.FirstName)
		argPos++
	}
	if updates.GetLastName() == "" {
		setClauses = append(setClauses, fmt.Sprintf("last_name = $%d", argPos))
		args = append(args, updates.LastName)
		argPos++
	}
	if updates.GetEmail() == "" {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argPos))
		args = append(args, updates.Email)
		argPos++
	}
	if updates.GetProfilePhotoURL() == "" {
		setClauses = append(setClauses, fmt.Sprintf("profile_photo = $%d", argPos))
		args = append(args, updates.ProfilePhotoURL)
		argPos++
	}

	if len(setClauses) == 0 {
		return nil, fmt.Errorf("%w: no fields provided to update", ErrUpdateUser)
	}

	// Finalize query
	query := fmt.Sprintf(`
		UPDATE users
		SET %s
		WHERE id = $%d
		RETURNING id, username, first_name, last_name, email, profile_photo, created_at
	`,
		strings.Join(setClauses, ", "),
		argPos,
	)

	args = append(args, id)

	var updated data.User

	err := db.DB.QueryRow(ctx, query, args...).Scan(
		&updated.ID,
		&updated.Username,
		&updated.FirstName,
		&updated.LastName,
		&updated.Email,
		&updated.ProfilePhotoURL,
		&updated.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) {
			if pgErr.Code == UniqueViolation {
				switch pgErr.ConstraintName {
				case "users_username_key":
					return nil, ErrUserNameExists
				case "users_email_key":
					return nil, ErrEmailExists
				default:
					return nil, fmt.Errorf("%w: constraint_name: %s", ErrUnknownUniqueConstraint, pgErr.ConstraintName)
				}
			}
			return nil, fmt.Errorf("%w: %w", ErrUpdateUser, pgErr)
		}
		return nil, fmt.Errorf("%w: %w", ErrUpdateUser, err)
	}

	return &updated, nil
}

func (db *Client) ProfileDelete(ctx context.Context, id string) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`

	tag, err := db.DB.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDeleteUser, err)
	}

	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

var ErrUnableToCreateUserTable = fmt.Errorf("failed to register user table")

func registerUser(ctx context.Context, db *pgx.Conn) error {
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		profile_photo TEXT,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
`
	if _, err := db.Exec(ctx, createUsersTable); err != nil {
		return fmt.Errorf("%w: %w", ErrUnableToCreateUserTable, err)
	}
	return nil
}
