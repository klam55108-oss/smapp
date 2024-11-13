package service

import (
	"context"
	"errors"
	"fmt"
	"smapp/post/config"
	"smapp/post/model"
	"smapp/post/repository"

	"github.com/google/uuid"
)

type Comment struct {
	commentRepository *repository.Comment
	postRepository    *repository.Post
}

func NewComment(commentRepository *repository.Comment, postRepository *repository.Post) *Comment {
	return &Comment{
		commentRepository: commentRepository,
		postRepository:    postRepository,
	}
}

func (svc *Comment) Create(ctx context.Context, postID, authorID uuid.UUID, body string) (uuid.UUID, error) {
	fail := func(err error) (uuid.UUID, error) {
		return uuid.Nil, fmt.Errorf("create comment: %w", err)
	}

	id, err := svc.commentRepository.Create(ctx, postID, authorID, body)
	if errors.Is(err, repository.ErrPostIDNotFound) {
		return uuid.Nil, fmt.Errorf("%w: %s", ErrPostNotFound, postID)
	}
	if err != nil {
		return fail(err)
	}

	return id, nil
}

var ErrCommentsPaginationLimitInvalid = errors.New("comments pagination limit invalid")

func (svc *Comment) GetPaginatedWithLikeCount(
	ctx context.Context, postID uuid.UUID, cursor model.Cursor, limit int,
) ([]model.Comment, *model.Cursor, error) {
	fail := func(err error) ([]model.Comment, *model.Cursor, error) {
		return nil, nil, fmt.Errorf("get comments: %w", err)
	}

	if limit < 1 || limit > config.CommentsPaginationLimit {
		return nil, nil, fmt.Errorf(
			"%w, should be in range: [1, %d]",
			ErrCommentsPaginationLimitInvalid, config.CommentsPaginationLimit,
		)
	}

	if err := svc.postRepository.CheckExists(ctx, postID); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("%w: %s", ErrPostNotFound, postID)
		}
		return fail(err)
	}

	comments, nextCursor, err := svc.commentRepository.GetPaginatedWithLikeCount(ctx, postID, cursor, limit)
	if err != nil {
		return fail(err)
	}

	return comments, nextCursor, nil
}
