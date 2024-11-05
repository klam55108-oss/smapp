package repository

import (
	"context"
	"database/sql"
	"fmt"
	"smapp/post/model"

	"github.com/google/uuid"
)

type Post struct {
	db *sql.DB
}

func NewPost(db *sql.DB) *Post {
	return &Post{db: db}
}

var ErrPostDoesNotExist = fmt.Errorf("post does not exist")

func (p *Post) Create(ctx context.Context, body, authorID string, images []model.ImageLocation) (string, error) {
	fail := func(err error) (string, error) {
		return "", fmt.Errorf("add post to db: %w", err)
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()

	id, err := uuid.NewRandom()
	if err != nil {
		return fail(err)
	}
	authorUUID, err := uuid.Parse(authorID)
	if err != nil {
		return fail(err)
	}
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO posts (id, body, author_id) VALUES (?, ?, ?)",
		id[:], body, authorUUID[:],
	)
	if err != nil {
		return fail(changeErrIfCtxDone(ctx, err))
	}

	for i, image := range images {
		imageID, err := uuid.NewRandom()
		if err != nil {
			return fail(err)
		}
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO images (id, post_id, position, s3_bucket, s3_key) VALUES (?, ?, ?, ?, ?)",
			imageID[:], id[:], i, image.Bucket, image.Key,
		)
		if err != nil {
			return fail(changeErrIfCtxDone(ctx, err))
		}
	}
	if err = tx.Commit(); err != nil {
		return fail(changeErrIfCtxDone(ctx, err))
	}
	return id.String(), nil
}

func (p *Post) CheckIDExists(ctx context.Context, id string) error {
	fail := func(err error) error {
		return fmt.Errorf("check if post exists in db: %w", err)
	}

	postUUID, err := uuid.Parse(id)
	if err != nil {
		return fail(err)
	}

	var exists bool
	err = p.db.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM posts WHERE id = ?)",
		postUUID[:],
	).Scan(&exists)
	if err != nil {
		return fail(err)
	}
	if !exists {
		return fail(fmt.Errorf("%w: %s", ErrPostDoesNotExist, id))
	}
	return nil
}
