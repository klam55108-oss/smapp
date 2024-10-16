package service

import (
	"smapp/user_service/repository"

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

func (svc *Auth) Signup(name, email, handle, password string) (string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	id, err := svc.userRepository.CreateUser(name, email, handle, string(passwordHash))
	if err != nil {
		return "", err
	}
	token, err := svc.jwtService.GenerateJWT(id.String())
	if err != nil {
		return "", err
	}
	return token, nil
}

func (svc *Auth) Login(identifier string, password []byte) (string, error) {
	id, passwordHash, err := svc.userRepository.GetAuthData(identifier)
	if err != nil {
		return "", err
	}
	err = bcrypt.CompareHashAndPassword(passwordHash, password)
	if err != nil {
		return "", err
	}
	token, err := svc.jwtService.GenerateJWT(id.String())
	if err != nil {
		return "", err
	}
	return token, nil
}
