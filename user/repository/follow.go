package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type Follow struct {
	db *sql.DB
}

func NewFollow(db *sql.DB) *Follow {
	return &Follow{db: db}
}

func (f *Follow) Create(ctx context.Context, followerID, followedID uuid.UUID) error {
	fail := func(err error) error {
		return fmt.Errorf("add follow to db: %w", err)
	}

	_, err := f.db.ExecContext(
		ctx,
		"INSERT INTO follows (follower_id, followed_id) VALUES (?, ?)",
		followerID[:], followedID[:],
	)
	var mysqlError *mysql.MySQLError
	if errors.As(err, &mysqlError) {
		if mysqlError.Number == 1452 {
			return ErrUserIDNotFound
		}
		if mysqlError.Number == 1062 {
			return ErrRecordExists
		}
	}
	if err != nil {
		return fail(err)
	}
	return nil
}

func (f *Follow) GetFollowed(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	fail := func(err error) ([]uuid.UUID, error) {
		return nil, fmt.Errorf("get followed from db: %w", err)
	}

	rows, err := f.db.QueryContext(
		ctx,
		"SELECT followed_id FROM follows WHERE follower_id = ?",
		userID[:],
	)
	if err != nil {
		return fail(err)
	}
	defer rows.Close()

	var followed []uuid.UUID
	for rows.Next() {
		var followedID uuid.UUID
		if err := rows.Scan(&followedID); err != nil {
			return fail(err)
		}
		followed = append(followed, followedID)
	}
	if err := rows.Err(); err != nil {
		return fail(err)
	}
	return followed, nil
}
