package service

import (
	"context"
	"smapp/post_service/repository"
)

type Post struct {
	postRepository *repository.Post
}

func NewPost(postRepository *repository.Post) *Post {
	return &Post{
		postRepository: postRepository,
	}
}

func (svc *Post) CreatePost(ctx context.Context, body string, author_id string, image_urls []string) (string, error) {
	// TODO: Check if the image_urls exist on S3, maybe in parallel

	id, err := svc.postRepository.CreatePost(ctx, body, author_id, image_urls)

	// TODO: Validate images on S3, and implement cleanup policies

	return id, err
}
