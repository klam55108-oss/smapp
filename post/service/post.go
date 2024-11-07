package service

import (
	"context"
	"fmt"
	"smapp/post/model"
	"smapp/post/repository"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	imagePB "smapp/common/grpc/image"
)

type Post struct {
	postRepository    *repository.Post
	imageClient       imagePB.ImageClient
	commentRepository *repository.Comment
	likeRepository    *repository.Like
}

func NewPost(
	postRepository *repository.Post, imageClient imagePB.ImageClient, commentRepository *repository.Comment,
	likeRepository *repository.Like,
) *Post {
	return &Post{
		postRepository:    postRepository,
		imageClient:       imageClient,
		commentRepository: commentRepository,
		likeRepository:    likeRepository,
	}
}

var ErrInvalidImage = fmt.Errorf("image does not exist or not accessible")

func (svc *Post) Create(
	ctx context.Context, body string, authorID string, images []model.ImageLocation,
) (string, error) {
	fail := func(err error) (string, error) {
		return "", fmt.Errorf("create post: %w", err)
	}

	for _, image := range images {
		_, err := svc.imageClient.CheckObjectExists(ctx, &imagePB.ObjectExistsRequest{
			Bucket: image.Bucket,
			Key:    image.Key,
		})

		if status.Code(err) != codes.OK {
			return fail(fmt.Errorf("%w: %s", ErrInvalidImage, err))
		}
	}

	id, err := svc.postRepository.Create(ctx, body, authorID, images)
	if err != nil {
		return fail(err)
	}

	// TODO: Implement expiration for certain tags

	return id, nil
}

func (svc *Post) Get(ctx context.Context, id string) (model.Post, int, int, error) {
	fail := func(err error) (model.Post, int, int, error) {
		return model.Post{}, 0, 0, fmt.Errorf("get post: %w", err)
	}

	post, err := svc.postRepository.Get(ctx, id)
	if err != nil {
		return fail(err)
	}

	commentCount, err := svc.commentRepository.GetCount(ctx, id)
	if err != nil {
		return fail(err)
	}

	likeCount, err := svc.likeRepository.GetCount(ctx, id)
	if err != nil {
		return fail(err)
	}

	return post, commentCount, likeCount, nil
}
