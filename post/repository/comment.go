package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"smapp/post/model"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type Comment struct {
	db *sql.DB
}

func NewComment(db *sql.DB) *Comment {
	return &Comment{db: db}
}

func (c *Comment) Create(ctx context.Context, postID, authorID uuid.UUID, body string) (uuid.UUID, error) {
	fail := func(err error) (uuid.UUID, error) {
		return uuid.Nil, fmt.Errorf("add comment to db: %w", err)
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

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO comments (id, post_id, author_id, body) VALUES (?, ?, ?, ?)",
		id[:], postID[:], authorID[:], body,
	)
	var mysqlError *mysql.MySQLError
	if errors.As(err, &mysqlError) && mysqlError.Number == 1452 {
		return uuid.Nil, ErrPostIDNotFound
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
		countID[:], postID[:],
	)
	if err != nil {
		return fail(changeErrIfCtxDone(ctx, err))
	}

	if err = tx.Commit(); err != nil {
		return fail(changeErrIfCtxDone(ctx, err))
	}
	return id, nil
}

func (c *Comment) CheckExists(ctx context.Context, id uuid.UUID) error {
	fail := func(err error) error {
		return fmt.Errorf("check if comment exists in db: %w", err)
	}

	var exists bool
	err := c.db.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM comments WHERE id = ?)",
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

func (c *Comment) GetPaginatedWithLikeCount(ctx context.Context, postID uuid.UUID, cursor model.Cursor, limit int) ([]model.Comment, *model.Cursor, error) {
	fail := func(err error) ([]model.Comment, *model.Cursor, error) {
		return nil, nil, fmt.Errorf("get comments from db: %w", err)
	}

	query := `
		SELECT c.id, c.author_id, c.body, c.created_at, IFNULL(lc.count, 0) 
		FROM comments c 
		LEFT JOIN likes_count lc ON lc.entity_type = 'comments' AND lc.entity_id = c.id 
		WHERE c.post_id = ? AND (c.created_at < ? OR (c.created_at = ? AND c.id > ?)) 
		ORDER BY c.created_at DESC, c.id 
		LIMIT ? 
	`
	rows, err := c.db.QueryContext(
		ctx,
		query,
		postID[:], cursor.LastLoadedTimestamp, cursor.LastLoadedTimestamp,
		cursor.LastLoadedID[:], limit+1,
	)
	if err != nil {
		return fail(err)
	}
	defer rows.Close()
	comments := make([]model.Comment, 0)
	for i := 0; i < limit && rows.Next(); i++ {
		var comment model.Comment
		var likeCount uint32
		if err = rows.Scan(&comment.ID, &comment.AuthorID, &comment.Body, &comment.CreatedAt, &likeCount); err != nil {
			return fail(err)
		}
		comment.PostID = postID
		comment.LikeCount = &likeCount
		comments = append(comments, comment)
	}
	if err = rows.Err(); err != nil {
		return fail(err)
	}
	var nextCursor *model.Cursor
	// Given that we have loaded limit+1 elements and iterated over at most limit elements, rows.Next() == false means its the last page.
	if rows.Next() {
		nextCursor = &model.Cursor{
			LastLoadedTimestamp: comments[len(comments)-1].CreatedAt,
			LastLoadedID:        comments[len(comments)-1].ID,
		}
	}
	return comments, nextCursor, nil
}

func (c *Comment) GetCount(ctx context.Context, postID uuid.UUID) (uint32, error) {
	fail := func(err error) (uint32, error) {
		return 0, fmt.Errorf("get comment count from db: %w", err)
	}

	var count uint32
	err := c.db.QueryRowContext(
		ctx,
		"SELECT count FROM comments_count WHERE post_id = ?",
		postID[:],
	).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return fail(err)
	}

	return count, nil
}
