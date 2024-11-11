package service

import (
	"context"
	"errors"
	"fmt"
	"smapp/post/repository"
)

type entityRepository interface {
	CheckExists(ctx context.Context, entityID string) error
}

type Like struct {
	likeRepository    *repository.Like
	entityRepository  entityRepository
	errEntityNotFound error
}

func NewPostLike(likeRepository *repository.Like, postRepository *repository.Post) *Like {
	return &Like{
		likeRepository:    likeRepository,
		entityRepository:  postRepository,
		errEntityNotFound: ErrPostNotFound,
	}
}

func NewCommentLike(likeRepository *repository.Like, commentRepository *repository.Comment) *Like {
	return &Like{
		likeRepository:    likeRepository,
		entityRepository:  commentRepository,
		errEntityNotFound: ErrCommentNotFound,
	}
}

var ErrLikeExists = errors.New("like already exists")

func (svc Like) Create(ctx context.Context, entityID, authorID string) error {
	fail := func(err error) error {
		return fmt.Errorf("create like: %w", err)
	}

	if err := svc.entityRepository.CheckExists(ctx, entityID); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return fmt.Errorf("%w: %s", svc.errEntityNotFound, entityID)
		}
		return fail(err)
	}

	err := svc.likeRepository.Create(ctx, entityID, authorID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordExists) {
			return ErrLikeExists
		}
		return fail(err)
	}

	return nil
}
