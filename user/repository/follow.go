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

var ErrFollowExists = errors.New("follow already exists")

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
			return fail(fmt.Errorf("%w: %s", ErrUserDoesNotExist, followedID))
		}
		if mysqlError.Number == 1062 {
			return fail(ErrFollowExists)
		}
	}
	if err != nil {
		return fail(err)
	}
	return nil
}
