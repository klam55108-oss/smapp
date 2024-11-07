package service

import (
	"context"
	"fmt"
	"smapp/post/repository"
)

type entityRepository interface {
	CheckExists(ctx context.Context, entityID string) error
}

type Like struct {
	likeRepository   *repository.Like
	entityRepository entityRepository
}

func NewPostLike(likeRepository *repository.Like, postRepository *repository.Post) *Like {
	return &Like{
		likeRepository:   likeRepository,
		entityRepository: postRepository,
	}
}

func NewCommentLike(likeRepository *repository.Like, commentRepository *repository.Comment) *Like {
	return &Like{
		likeRepository:   likeRepository,
		entityRepository: commentRepository,
	}
}

func (svc Like) Create(ctx context.Context, entityID, authorID string) error {
	fail := func(err error) error {
		return fmt.Errorf("create like: %w", err)
	}

	if err := svc.entityRepository.CheckExists(ctx, entityID); err != nil {
		return fail(err)
	}

	err := svc.likeRepository.Create(ctx, entityID, authorID)
	if err != nil {
		return fail(err)
	}

	return nil
}
