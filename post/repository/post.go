package repository

import (
	"context"
	"database/sql"
	"errors"
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

func (p *Post) CheckExists(ctx context.Context, id string) error {
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

func (p *Post) Get(ctx context.Context, id string) (model.Post, error) {
	fail := func(err error) (model.Post, error) {
		return model.Post{}, fmt.Errorf("get post from db: %w", err)
	}

	postUUID, err := uuid.Parse(id)
	if err != nil {
		return fail(err)
	}

	var post model.Post
	post.ID = id

	var authorID []byte
	err = p.db.QueryRowContext(
		ctx,
		"SELECT author_id, body, created_at FROM posts WHERE id = ?",
		postUUID[:],
	).Scan(&authorID, &post.Body, &post.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return fail(fmt.Errorf("%w: %s", ErrPostDoesNotExist, id))
	}
	if err != nil {
		return fail(err)
	}

	authorUUID, err := uuid.FromBytes(authorID)
	if err != nil {
		return fail(err)
	}
	post.AuthorID = authorUUID.String()

	rows, err := p.db.QueryContext(
		ctx,
		"SELECT s3_bucket, s3_key FROM images WHERE post_id = ? ORDER BY position",
		postUUID[:],
	)
	if err != nil {
		return fail(err)
	}
	defer rows.Close()
	post.Images = make([]model.ImageLocation, 0)
	for rows.Next() {
		var image model.ImageLocation
		if err = rows.Scan(&image.Bucket, &image.Key); err != nil {
			return fail(err)
		}
		post.Images = append(post.Images, image)
	}
	if err = rows.Err(); err != nil {
		return fail(err)
	}

	return post, nil
}
