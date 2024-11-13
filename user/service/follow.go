package service

import (
	"context"
	"errors"
	"fmt"
	"smapp/user/repository"

	"github.com/google/uuid"
)

type Follow struct {
	followRepository *repository.Follow
}

func NewFollow(followRepository *repository.Follow) *Follow {
	return &Follow{
		followRepository: followRepository,
	}
}

var (
	ErrSelfFollow   = errors.New("cannot follow self")
	ErrFollowExists = errors.New("follow already exists")
)

func (svc *Follow) Create(ctx context.Context, followerID, followedID uuid.UUID) error {
	fail := func(err error) error {
		return fmt.Errorf("create follow: %w", err)
	}
	if followerID == followedID {
		return ErrSelfFollow
	}
	err := svc.followRepository.Create(ctx, followerID, followedID)
	if errors.Is(err, repository.ErrUserIDNotFound) {
		return fmt.Errorf("%w: %s", ErrUserNotFound, followedID)
	}
	if errors.Is(err, repository.ErrRecordExists) {
		return ErrFollowExists
	}
	if err != nil {
		return fail(err)
	}
	return nil
}

func (svc *Follow) GetFollowed(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	fail := func(err error) ([]uuid.UUID, error) {
		return nil, fmt.Errorf("get followed: %w", err)
	}
	followed, err := svc.followRepository.GetFollowed(ctx, userID)
	if err != nil {
		return fail(err)
	}
	return followed, nil
}
