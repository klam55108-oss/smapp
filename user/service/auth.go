package service

import (
	"context"
	"fmt"
	"smapp/user/repository"

	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	userRepository *repository.User
	jwtService     *JWT
}

func NewAuth(userRepository *repository.User, jwtService *JWT) *Auth {
	return &Auth{
		userRepository: userRepository,
		jwtService:     jwtService,
	}
}

func (svc *Auth) Signup(ctx context.Context, name, email, handle, password string) (string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("signup: %w", err)
	}
	id, err := svc.userRepository.Create(ctx, name, email, handle, string(passwordHash))
	if err != nil {
		return "", fmt.Errorf("signup: %w", err)
	}
	token, err := svc.jwtService.GenerateJWT(id)
	if err != nil {
		return "", fmt.Errorf("signup: %w", err)
	}
	return token, nil
}

func (svc *Auth) Login(ctx context.Context, identifier string, password []byte) (string, error) {
	id, passwordHash, err := svc.userRepository.GetAuthData(ctx, identifier)
	if err != nil {
		return "", fmt.Errorf("login: %w", err)
	}
	err = bcrypt.CompareHashAndPassword(passwordHash, password)
	if err != nil {
		return "", fmt.Errorf("login: %w", err)
	}
	token, err := svc.jwtService.GenerateJWT(id)
	if err != nil {
		return "", fmt.Errorf("login: %w", err)
	}
	return token, nil
}
