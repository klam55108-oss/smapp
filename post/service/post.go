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
	postRepository *repository.Post
	imageClient    imagePB.ImageClient
}

func NewPost(postRepository *repository.Post, imageClient imagePB.ImageClient) *Post {
	return &Post{
		postRepository: postRepository,
		imageClient:    imageClient,
	}
}

var ErrInvalidImage = fmt.Errorf("image does not exist or not accessible")

func (svc *Post) Create(
	ctx context.Context, body string, author_id string, images []model.ImageLocation,
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

	id, err := svc.postRepository.Create(ctx, body, author_id, images)
	if err != nil {
		return fail(err)
	}

	// TODO: Implement expiration for certain tags

	return id, nil
}
