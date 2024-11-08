package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type Comment struct {
	db *sql.DB
}

func NewComment(db *sql.DB) *Comment {
	return &Comment{db: db}
}

var ErrCommentDoesNotExist = fmt.Errorf("comment does not exist")

func (c *Comment) Create(ctx context.Context, postID, authorID, body string) (string, error) {
	fail := func(err error) (string, error) {
		return "", fmt.Errorf("add comment to db: %w", err)
	}

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()

	id, err := uuid.NewRandom()
	if err != nil {
		return fail(err)
	}
	postUUID, err := uuid.Parse(postID)
	if err != nil {
		return fail(err)
	}
	authorUUID, err := uuid.Parse(authorID)
	if err != nil {
		return fail(err)
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO comments (id, post_id, author_id, body) VALUES (?, ?, ?, ?)",
		id[:], postUUID[:], authorUUID[:], body,
	)
	var mysqlError *mysql.MySQLError
	if errors.As(err, &mysqlError) && mysqlError.Number == 1452 {
		return fail(fmt.Errorf("%w: %s", ErrPostDoesNotExist, postID))
	}
	if err != nil {
		return fail(changeErrIfCtxDone(ctx, err))
	}

	countID, err := uuid.NewRandom()
	if err != nil {
		return fail(err)
	}
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO comments_count (id, post_id, count) VALUES (?, ?, 1) ON DUPLICATE KEY UPDATE count = count + 1",
		countID[:], postUUID[:],
	)
	if err != nil {
		return fail(changeErrIfCtxDone(ctx, err))
	}

	if err = tx.Commit(); err != nil {
		return fail(changeErrIfCtxDone(ctx, err))
	}
	return id.String(), nil
}

func (c *Comment) CheckExists(ctx context.Context, id string) error {
	fail := func(err error) error {
		return fmt.Errorf("check if comment exists in db: %w", err)
	}

	commentUUID, err := uuid.Parse(id)
	if err != nil {
		return fail(err)
	}

	var exists bool
	err = c.db.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM comments WHERE id = ?)",
		commentUUID[:],
	).Scan(&exists)
	if err != nil {
		return fail(err)
	}
	if !exists {
		return fail(fmt.Errorf("%w: %s", ErrCommentDoesNotExist, id))
	}
	return nil
}

func (c *Comment) GetCount(ctx context.Context, postID string) (int, error) {
	fail := func(err error) (int, error) {
		return 0, fmt.Errorf("get comment count from db: %w", err)
	}

	postUUID, err := uuid.Parse(postID)
	if err != nil {
		return fail(err)
	}

	var count int
	err = c.db.QueryRowContext(
		ctx,
		"SELECT count FROM comments_count WHERE post_id = ?",
		postUUID[:],
	).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return fail(err)
	}

	return count, nil
}
