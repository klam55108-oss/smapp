package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"smapp/post/model"

	"github.com/google/uuid"
)

type Like struct {
	db         *sql.DB
	entityType model.EntityType
}

func NewPostLike(db *sql.DB) *Like {
	return &Like{db: db, entityType: model.PostType}
}

func NewCommentLike(db *sql.DB) *Like {
	return &Like{db: db, entityType: model.CommentType}
}

var ErrLikeExists = errors.New("like already exists")

func (l *Like) Create(ctx context.Context, entityID, authorID string) error {
	fail := func(err error) error {
		return fmt.Errorf("add like to db: %w", err)
	}

	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()

	id, err := uuid.NewRandom()
	if err != nil {
		return fail(err)
	}
	entityUUID, err := uuid.Parse(entityID)
	if err != nil {
		return fail(err)
	}
	authorUUID, err := uuid.Parse(authorID)
	if err != nil {
		return fail(err)
	}

	result, err := tx.ExecContext(
		ctx,
		"INSERT IGNORE INTO likes (id, entity_type, entity_id, author_id) VALUES (?, ?, ?, ?)",
		id[:], l.entityType, entityUUID[:], authorUUID[:],
	)
	if err != nil {
		return fail(changeErrIfCtxDone(ctx, err))
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fail(err)
	}
	if rowsAffected == 0 {
		return fail(ErrLikeExists)
	}

	countID, err := uuid.NewRandom()
	if err != nil {
		return fail(err)
	}
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO likes_count (id, entity_type, entity_id, count) VALUES (?, ?, ?, 1) ON DUPLICATE KEY UPDATE count = count + 1",
		countID[:], l.entityType, entityUUID[:],
	)
	if err != nil {
		return fail(changeErrIfCtxDone(ctx, err))
	}

	if err = tx.Commit(); err != nil {
		return fail(changeErrIfCtxDone(ctx, err))
	}
	return nil
}

func (l *Like) GetCount(ctx context.Context, entityID string) (int, error) {
	fail := func(err error) (int, error) {
		return 0, fmt.Errorf("get like count from db: %w", err)
	}

	entityUUID, err := uuid.Parse(entityID)
	if err != nil {
		return fail(err)
	}

	var count int
	err = l.db.QueryRowContext(
		ctx,
		"SELECT count FROM likes_count WHERE entity_type = ? AND entity_id = ?",
		l.entityType, entityUUID[:],
	).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return fail(err)
	}

	return count, nil
}
