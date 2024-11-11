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

type Post struct {
	db *sql.DB
}

func NewPost(db *sql.DB) *Post {
	return &Post{db: db}
}

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
		return ErrRecordNotFound
	}
	return nil
}

func (p *Post) getImagesByPostID(ctx context.Context, postUUID uuid.UUID) ([]model.ImageLocation, error) {
	fail := func(err error) ([]model.ImageLocation, error) {
		return nil, fmt.Errorf("get images by post id from db: %w", err)
	}
	rows, err := p.db.QueryContext(
		ctx,
		"SELECT s3_bucket, s3_key FROM images WHERE post_id = ? ORDER BY position",
		postUUID[:],
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

	var authorIDBytes []byte
	err = p.db.QueryRowContext(
		ctx,
		"SELECT author_id, body, created_at FROM posts WHERE id = ?",
		postUUID[:],
	).Scan(&authorIDBytes, &post.Body, &post.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Post{}, ErrRecordNotFound
	}
	if err != nil {
		return fail(err)
	}

	authorUUID, err := uuid.FromBytes(authorIDBytes)
	if err != nil {
		return fail(err)
	}
	post.AuthorID = authorUUID.String()

	post.Images, err = p.getImagesByPostID(ctx, postUUID)
	if err != nil {
		return fail(err)
	}

	return post, nil
}

func (p *Post) GetWithCountsByUserIDs(
	ctx context.Context, userIDs []string, cursor model.Cursor, limit int,
) ([]model.Post, *model.Cursor, error) {
	fail := func(err error) ([]model.Post, *model.Cursor, error) {
		return nil, nil, fmt.Errorf("get posts by user ids from db: %w", err)
	}

	hexUUIDs := make([]string, len(userIDs))
	for i, userID := range userIDs {
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			return fail(err)
		}
		hexUUIDs[i] = fmt.Sprintf("X'%x'", userUUID[:])
	}

	lastLoadedUUID, err := uuid.Parse(cursor.LastLoadedID)
	if err != nil {
		return fail(err)
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.author_id, p.body, p.created_at, IFNULL(cc.count, 0), IFNULL(lc.count, 0) 
		FROM posts p 
		LEFT JOIN comments_count cc ON cc.post_id = p.id 
		LEFT JOIN likes_count lc ON lc.entity_type = 'posts' AND lc.entity_id = p.id 
		WHERE p.author_id IN (%s) AND (p.created_at < ? OR (p.created_at = ? AND p.id > ?)) 
		ORDER BY p.created_at DESC, p.id 
		LIMIT ? 
	`, strings.Join(hexUUIDs, ","))
	rows, err := p.db.QueryContext(
		ctx, query,
		cursor.LastLoadedTimestamp, cursor.LastLoadedTimestamp, lastLoadedUUID[:], limit+1,
	)
	if err != nil {
		return fail(err)
	}
	defer rows.Close()

	posts := make([]model.Post, 0)
	for i := 0; i < limit && rows.Next(); i++ {
		var post model.Post
		var postIDBytes, authorIDBytes []byte
		var commentCount, likeCount uint32
		if err = rows.Scan(&postIDBytes, &authorIDBytes, &post.Body, &post.CreatedAt, &commentCount, &likeCount); err != nil {
			return fail(err)
		}
		postUUID, err := uuid.FromBytes(postIDBytes)
		if err != nil {
			return fail(err)
		}
		post.ID = postUUID.String()
		authorUUID, err := uuid.FromBytes(authorIDBytes)
		if err != nil {
			return fail(err)
		}
		post.AuthorID = authorUUID.String()
		post.CommentCount = &commentCount
		post.LikeCount = &likeCount
		post.Images, err = p.getImagesByPostID(ctx, postUUID)
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
