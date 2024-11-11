package service

import (
	"context"
	"errors"
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
			return "", fmt.Errorf("%w: %s", ErrInvalidImage, err)
		}
	}

	id, err := svc.postRepository.Create(ctx, body, authorID, images)
	if err != nil {
		return fail(err)
	}

	// TODO: Implement expiration for certain tags

	return id, nil
}

// TODO: implement WithLikeCount/WithCommentCount options
func (svc *Post) GetWithCounts(ctx context.Context, id string) (model.Post, error) {
	fail := func(err error) (model.Post, error) {
		return model.Post{}, fmt.Errorf("get post: %w", err)
	}

	post, err := svc.postRepository.Get(ctx, id)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return model.Post{}, fmt.Errorf("%w: %s", ErrPostNotFound, id)
	}
	if err != nil {
		return fail(err)
	}

	commentCount, err := svc.commentRepository.GetCount(ctx, id)
	if err != nil {
		return fail(err)
	}
	post.CommentCount = &commentCount

	likeCount, err := svc.likeRepository.GetCount(ctx, id)
	if err != nil {
		return fail(err)
	}
	post.LikeCount = &likeCount

	return post, nil
}
