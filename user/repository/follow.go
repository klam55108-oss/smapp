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

func (f *Follow) Create(ctx context.Context, followerID, followedID string) error {
	fail := func(err error) error {
		return fmt.Errorf("add follow to db: %w", err)
	}

	followerUUID, err := uuid.Parse(followerID)
	if err != nil {
		return fail(err)
	}
	followedUUID, err := uuid.Parse(followedID)
	if err != nil {
		return fail(err)
	}
	_, err = f.db.ExecContext(
		ctx,
		"INSERT INTO follows (follower_id, followed_id) VALUES (?, ?)",
		followerUUID[:], followedUUID[:],
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

func (f *Follow) GetFollowed(ctx context.Context, userID string) ([]string, error) {
	fail := func(err error) ([]string, error) {
		return nil, fmt.Errorf("get followed from db: %w", err)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fail(err)
	}
	rows, err := f.db.QueryContext(
		ctx,
		"SELECT followed_id FROM follows WHERE follower_id = ?",
		userUUID[:],
	)
	if err != nil {
		return fail(err)
	}
	defer rows.Close()

	var followed []string
	for rows.Next() {
		var followedIDBytes []byte
		if err := rows.Scan(&followedIDBytes); err != nil {
			return fail(err)
		}
		followedUUID, err := uuid.FromBytes(followedIDBytes)
		if err != nil {
			return fail(err)
		}
		followed = append(followed, followedUUID.String())
	}
	if err := rows.Err(); err != nil {
		return fail(err)
	}
	return followed, nil
}
