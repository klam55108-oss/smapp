package service

import (
	"context"
	"fmt"
	"smapp/post/repository"

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

var ErrImageNotFound = fmt.Errorf("image not found")

func (svc *Post) Create(ctx context.Context, body string, author_id string, image_urls []string) (string, error) {
	fail := func(err error) (string, error) {
		return "", fmt.Errorf("create post: %w", err)
	}

	// TODO: Check if the image_urls exist on S3, maybe in parallel

	for _, url := range image_urls {
		exists, err := svc.imageClient.VerifyURL(ctx, &imagePB.URL{Url: url})
		if err != nil {
			return fail(err)
		}
		if !exists.Exists {
			return fail(fmt.Errorf("%w: %s", ErrImageNotFound, url))
		}
	}

	id, err := svc.postRepository.Create(ctx, body, author_id, image_urls)
	if err != nil {
		return fail(err)
	}

	// TODO: Validate images on S3, and implement cleanup policies

	return id, nil
}
