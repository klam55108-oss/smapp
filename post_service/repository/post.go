package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type Post struct {
	db *sql.DB
}

func (p *Post) Close() error {
	return p.db.Close()
}

func NewPost(db *sql.DB) *Post {
	return &Post{db: db}
}

// tx operations may return sql.ErrTxDone if the context is done. Return a context error instead for clarity in the service layer
func changeErrIfCtxDone(ctx context.Context, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil && errors.Is(err, sql.ErrTxDone) {
		return ctxErr
	}
	return err
}

func (p *Post) CreatePost(ctx context.Context, body, author_id string, image_urls []string) (string, error) {
	fail := func(err error) (string, error) {
		return "", fmt.Errorf("create post: %w", err)
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
	author_uuid, err := uuid.Parse(author_id)
	if err != nil {
		return fail(err)
	}
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO posts (id, body, author_id) VALUES (?, ?, ?)",
		id[:], body, author_uuid[:],
	)
	if err != nil {
		return fail(changeErrIfCtxDone(ctx, err))
	}

	for i, url := range image_urls {
		image_id, err := uuid.NewRandom()
		if err != nil {
			return fail(err)
		}
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO images (id, post_id, position, url) VALUES (?, ?, ?, ?)",
			image_id[:], id[:], i, url,
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
