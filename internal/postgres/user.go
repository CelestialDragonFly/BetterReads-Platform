package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/celestialdragonfly/betterreads/internal/data"
	"github.com/celestialdragonfly/betterreads/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrFollowUser           = errors.New("FollowUser: failed to follow user")
	ErrSelfFollow           = errors.New("FollowUser: cannot follow self")
	ErrUnableToCreateFollow = errors.New("failed to register ensure follows table")
	ErrUnfollowUser         = errors.New("failed to unfollow user")
)

func (db *Client) GetUserByID(ctx context.Context, id string) (*data.User, error) {
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

func (db *Client) FollowUser(ctx context.Context, followerID, followeeID string) error {
	if followerID == followeeID {
		return ErrSelfFollow
	}
	query := `
		INSERT INTO follows (follower_id, followee_id)
		VALUES ($1, $2)
	`

	_, err := db.DB.Exec(ctx, query, followerID, followeeID)
	if err != nil {
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) {
			if pgErr.Code == UniqueViolation {
				logger.Warn("Already following user", "follower_id", followerID, "followee_id", followeeID)
				return nil
			}
		}
		return fmt.Errorf("%w: %w", ErrFollowUser, err)
	}

	return nil
}

func (db *Client) UnfollowUser(ctx context.Context, followerID, followeeID string) error {
	if followerID == followeeID {
		return ErrSelfFollow
	}

	query := `
        DELETE FROM follows 
        WHERE follower_id = $1 AND followee_id = $2
    `

	_, err := db.DB.Exec(ctx, query, followerID, followeeID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnfollowUser, err)
	}

	return nil
}

func registerFollows(ctx context.Context, db *pgx.Conn) error {
	createFollowsTable := `
	CREATE TABLE IF NOT EXISTS follows (
		follower_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		followee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		PRIMARY KEY (follower_id, followee_id)
	);
	`
	if _, err := db.Exec(ctx, createFollowsTable); err != nil {
		return fmt.Errorf("%w: %w", ErrUnableToCreateFollow, err)
	}
	return nil
}
