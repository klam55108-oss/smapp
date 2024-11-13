package service

import (
	"context"
	"errors"
	"fmt"
	"smapp/post/config"
	"smapp/post/model"
	"smapp/post/repository"

	imagePB "smapp/common/grpc/image"
	userPB "smapp/common/grpc/user"

	"github.com/google/uuid"
)

type Post struct {
	postRepository    *repository.Post
	commentRepository *repository.Comment
	likeRepository    *repository.Like
	userClient        userPB.UserClient
	imageClient       imagePB.ImageClient
}

func NewPost(
	postRepository *repository.Post, commentRepository *repository.Comment, likeRepository *repository.Like,
	userClient userPB.UserClient, imageClient imagePB.ImageClient,
) *Post {
	return &Post{
		postRepository:    postRepository,
		commentRepository: commentRepository,
		likeRepository:    likeRepository,
		imageClient:       imageClient,
		userClient:        userClient,
	}
}

var ErrInvalidImage = fmt.Errorf("image does not exist or not accessible")

func (svc *Post) Create(
	ctx context.Context, body string, authorID uuid.UUID, images []model.ImageLocation,
) (uuid.UUID, error) {
	fail := func(err error) (uuid.UUID, error) {
		return uuid.Nil, fmt.Errorf("create post: %w", err)
	}

	for _, image := range images {
		_, err := svc.imageClient.CheckObjectExists(ctx, &imagePB.ObjectExistsRequest{
			Bucket: image.Bucket,
			Key:    image.Key,
		})

		if err != nil {
			return uuid.Nil, fmt.Errorf("%w: %w", ErrInvalidImage, err)
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
func (svc *Post) GetWithCounts(ctx context.Context, id uuid.UUID) (model.Post, error) {
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

var ErrPostsPaginationLimitInvalid = errors.New("posts pagination limit invalid")

func (svc *Post) GetFeed(
	ctx context.Context, authorID uuid.UUID, cursor model.Cursor, limit int,
) ([]model.Post, *model.Cursor, error) {
	fail := func(err error) ([]model.Post, *model.Cursor, error) {
		return nil, nil, fmt.Errorf("get feed: %w", err)
	}

	if limit < 1 || limit > config.CommentsPaginationLimit {
		return nil, nil, fmt.Errorf(
			"%w, should be in range: [1, %d]",
			ErrPostsPaginationLimitInvalid, config.CommentsPaginationLimit,
		)
	}

	followed, err := svc.userClient.GetFollowed(ctx, &userPB.GetFollowedRequest{UserId: authorID[:]})
	if err != nil {
		return fail(err)
	}
	if len(followed.UserIds) == 0 {
		return []model.Post{}, nil, nil
	}
	userIDs := make([]uuid.UUID, len(followed.UserIds))
	for i, userID := range followed.UserIds {
		userIDs[i], err = uuid.FromBytes(userID)
		if err != nil {
			return fail(err)
		}
	}

	posts, nextCursor, err := svc.postRepository.GetWithCountsByUserIDs(ctx, userIDs, cursor, limit)
	if err != nil {
		return fail(err)
	}

	return posts, nextCursor, nil
}
