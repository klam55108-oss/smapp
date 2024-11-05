package service

import (
	"context"
	"fmt"
	"smapp/post/repository"
)

type Like struct {
	likeRepository    *repository.Like
	postRepository    *repository.Post
	commentRepository *repository.Comment
}

func NewLike(
	likeRepository *repository.Like, postRepository *repository.Post, commentRepository *repository.Comment,
) *Like {
	return &Like{
		likeRepository:    likeRepository,
		postRepository:    postRepository,
		commentRepository: commentRepository,
	}
}

func (svc *Like) Create(ctx context.Context, entityType, entityID, authorID string) error {
	fail := func(err error) error {
		return fmt.Errorf("create like: %w", err)
	}

	if entityType == "posts" {
		if err := svc.postRepository.CheckIDExists(ctx, entityID); err != nil {
			return fail(err)
		}
	} else if entityType == "comments" {
		if err := svc.commentRepository.CheckIDExists(ctx, entityID); err != nil {
			return fail(err)
		}
	} else {
		return fail(fmt.Errorf("unknown entity type: %s", entityType))
	}

	err := svc.likeRepository.Create(ctx, entityType, entityID, authorID)
	if err != nil {
		return fail(err)
	}

	return nil
}
