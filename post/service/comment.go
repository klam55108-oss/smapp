package service

import (
	"context"
	"fmt"
	"smapp/post/model"
	"smapp/post/repository"
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

func (svc *Comment) Create(ctx context.Context, postID, authorID, body string) (string, error) {
	fail := func(err error) (string, error) {
		return "", fmt.Errorf("create comment: %w", err)
	}

	id, err := svc.commentRepository.Create(ctx, postID, authorID, body)
	if err != nil {
		return fail(err)
	}

	return id, nil
}

func (svc *Comment) GetPaginatedWithLikeCount(
	ctx context.Context, postID string, cursor model.Cursor, limit int,
) ([]model.Comment, *model.Cursor, error) {
	fail := func(err error) ([]model.Comment, *model.Cursor, error) {
		return nil, nil, fmt.Errorf("get comments: %w", err)
	}

	if err := svc.postRepository.CheckExists(ctx, postID); err != nil {
		return fail(err)
	}

	comments, nextCursor, err := svc.commentRepository.GetPaginatedWithLikeCount(ctx, postID, cursor, limit)
	if err != nil {
		return fail(err)
	}

	return comments, nextCursor, nil
}
