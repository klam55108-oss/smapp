package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	secret         []byte
	expirationTime time.Duration
}

func NewJWT(secret []byte, expirationTime time.Duration) *JWT {
	return &JWT{
		secret:         secret,
		expirationTime: expirationTime,
	}
}

func (svc *JWT) GenerateJWT(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(svc.expirationTime).Unix(),
	})
	return token.SignedString(svc.secret)
}

func (svc *JWT) ParseJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return svc.secret, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if sub, ok := claims["sub"].(string); ok {
			return sub, nil
		}
		return "", errors.New("expected sub claim to be present and to be a string")
	}
	return "", errors.New("expected claims to be jwt.MapClaims")
}
