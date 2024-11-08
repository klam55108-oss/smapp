package service

import (
	"context"
	"errors"
	"fmt"
	"smapp/user/repository"
)

type Follow struct {
	followRepository *repository.Follow
}

func NewFollow(followRepository *repository.Follow) *Follow {
	return &Follow{
		followRepository: followRepository,
	}
}

var ErrSelfFollow = errors.New("cannot follow self")

func (svc *Follow) Create(ctx context.Context, followerID, followedID string) error {
	fail := func(err error) error {
		return fmt.Errorf("create follow: %w", err)
	}
	if followerID == followedID {
		return fail(ErrSelfFollow)
	}
	err := svc.followRepository.Create(ctx, followerID, followedID)
	if err != nil {
		return fail(err)
	}
	return nil
}
