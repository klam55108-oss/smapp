package service

import (
	"context"
	"errors"
	"fmt"
	"smapp/user/repository"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	userRepository *repository.User
	jwtService     *JWT
}

func NewUser(userRepository *repository.User, jwtService *JWT) *User {
	return &User{
		userRepository: userRepository,
		jwtService:     jwtService,
	}
}

var (
	ErrEmailExists  = errors.New("email already exists")
	ErrHandleExists = errors.New("handle already exists")
)

func (svc *User) Signup(ctx context.Context, name, email, handle, password string) (string, error) {
	fail := func(err error) (string, error) {
		return "", fmt.Errorf("signup: %w", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fail(err)
	}
	id, err := svc.userRepository.Create(ctx, name, email, handle, string(passwordHash))
	if errors.Is(err, repository.ErrEmailExists) {
		return "", ErrEmailExists
	}
	if errors.Is(err, repository.ErrHandleExists) {
		return "", ErrHandleExists
	}
	if err != nil {
		return fail(err)
	}
	token, err := svc.jwtService.GenerateJWT(id)
	if err != nil {
		return fail(err)
	}
	return token, nil
}

func (svc *User) Login(ctx context.Context, identifier string, password []byte) (string, error) {
	fail := func(err error) (string, error) {
		return "", fmt.Errorf("login: %w", err)
	}

	id, passwordHash, err := svc.userRepository.GetAuthData(ctx, identifier)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return "", ErrUserNotFound
	}
	if err != nil {
		return fail(err)
	}
	err = bcrypt.CompareHashAndPassword(passwordHash, password)
	if err != nil {
		return fail(err)
	}
	token, err := svc.jwtService.GenerateJWT(id)
	if err != nil {
		return fail(err)
	}
	return token, nil
}
