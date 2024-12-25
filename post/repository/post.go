package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"smapp/post/model"
	"strings"

	"github.com/google/uuid"
)

//go:generate mockgen -destination mocks/post.go -package mocks . Post

type Post interface {
	Create(ctx context.Context, body string, authorID uuid.UUID, images []model.ImageLocation) (uuid.UUID, error)
	CheckExists(ctx context.Context, id uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID) (model.Post, error)
	GetWithCountsByUserIDs(ctx context.Context, userIDs []uuid.UUID, cursor model.Cursor, limit int) ([]model.Post, *model.Cursor, error)
}

type DefaultPost struct {
	db *sql.DB
}

func NewDefaultPost(db *sql.DB) *DefaultPost {
	return &DefaultPost{db: db}
}

func (p *DefaultPost) Create(ctx context.Context, body string, authorID uuid.UUID, images []model.ImageLocation) (uuid.UUID, error) {
	fail := func(err error) (uuid.UUID, error) {
		return uuid.Nil, fmt.Errorf("add post to db: %w", err)
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
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO posts (id, body, author_id) VALUES (?, ?, ?)",
		id[:], body, authorID[:],
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
	return id, nil
}

func (p *DefaultPost) CheckExists(ctx context.Context, id uuid.UUID) error {
	fail := func(err error) error {
		return fmt.Errorf("check if post exists in db: %w", err)
	}

	var exists bool
	err := p.db.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM posts WHERE id = ?)",
		id[:],
	).Scan(&exists)
	if err != nil {
		return fail(err)
	}
	if !exists {
		return ErrRecordNotFound
	}
	return nil
}

func (p *DefaultPost) getImagesByPostID(ctx context.Context, postID uuid.UUID) ([]model.ImageLocation, error) {
	fail := func(err error) ([]model.ImageLocation, error) {
		return nil, fmt.Errorf("get images by post id from db: %w", err)
	}
	rows, err := p.db.QueryContext(
		ctx,
		"SELECT s3_bucket, s3_key FROM images WHERE post_id = ? ORDER BY position",
		postID[:],
	)
	if err != nil {
		return fail(err)
	}
	defer rows.Close()
	images := make([]model.ImageLocation, 0)
	for rows.Next() {
		var image model.ImageLocation
		if err = rows.Scan(&image.Bucket, &image.Key); err != nil {
			return fail(err)
		}
		images = append(images, image)
	}
	if err = rows.Err(); err != nil {
		return fail(err)
	}
	return images, nil
}

func (p *DefaultPost) Get(ctx context.Context, id uuid.UUID) (model.Post, error) {
	fail := func(err error) (model.Post, error) {
		return model.Post{}, fmt.Errorf("get post from db: %w", err)
	}

	var post model.Post
	post.ID = id
	err := p.db.QueryRowContext(
		ctx,
		"SELECT author_id, body, created_at FROM posts WHERE id = ?",
		id[:],
	).Scan(&post.AuthorID, &post.Body, &post.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Post{}, ErrRecordNotFound
	}
	if err != nil {
		return fail(err)
	}

	post.Images, err = p.getImagesByPostID(ctx, id)
	if err != nil {
		return fail(err)
	}

	return post, nil
}

func (p *DefaultPost) GetWithCountsByUserIDs(
	ctx context.Context, userIDs []uuid.UUID, cursor model.Cursor, limit int,
) ([]model.Post, *model.Cursor, error) {
	fail := func(err error) ([]model.Post, *model.Cursor, error) {
		return nil, nil, fmt.Errorf("get posts by user ids from db: %w", err)
	}

	hexUserIDs := make([]string, len(userIDs))
	for i, userID := range userIDs {
		hexUserIDs[i] = fmt.Sprintf("X'%x'", userID[:])
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.author_id, p.body, p.created_at, IFNULL(cc.count, 0), IFNULL(lc.count, 0) 
		FROM posts p 
		LEFT JOIN comments_count cc ON cc.post_id = p.id 
		LEFT JOIN likes_count lc ON lc.entity_type = 'posts' AND lc.entity_id = p.id 
		WHERE p.author_id IN (%s) AND (p.created_at < ? OR (p.created_at = ? AND p.id > ?)) 
		ORDER BY p.created_at DESC, p.id 
		LIMIT ? 
	`, strings.Join(hexUserIDs, ","))
	rows, err := p.db.QueryContext(
		ctx, query,
		cursor.LastLoadedTimestamp, cursor.LastLoadedTimestamp, cursor.LastLoadedID[:], limit+1,
	)
	if err != nil {
		return fail(err)
	}
	defer rows.Close()

	posts := make([]model.Post, 0)
	for i := 0; i < limit && rows.Next(); i++ {
		var post model.Post
		var commentCount, likeCount uint32
		if err = rows.Scan(&post.ID, &post.AuthorID, &post.Body, &post.CreatedAt, &commentCount, &likeCount); err != nil {
			return fail(err)
		}
		post.CommentCount = &commentCount
		post.LikeCount = &likeCount
		post.Images, err = p.getImagesByPostID(ctx, post.ID)
		if err != nil {
			return fail(err)
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		return fail(err)
	}

	var nextCursor *model.Cursor
	// Given that we have loaded limit+1 elements and iterated over at most limit elements, rows.Next() == false means its the last page.
	if rows.Next() {
		nextCursor = &model.Cursor{
			LastLoadedTimestamp: posts[len(posts)-1].CreatedAt,
			LastLoadedID:        posts[len(posts)-1].ID,
		}
	}

	return posts, nextCursor, nil
}
