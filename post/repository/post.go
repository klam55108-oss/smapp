package repository

import (
	"context"
	"database/sql"
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

func (p *Post) Create(ctx context.Context, body, author_id string, image_urls []string) (string, error) {
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
