package service

import (
	"context"
	"fmt"
	"smapp/post/repository"
)

type Comment struct {
	commentRepository *repository.Comment
}

func NewComment(commentRepository *repository.Comment) *Comment {
	return &Comment{
		commentRepository: commentRepository,
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
