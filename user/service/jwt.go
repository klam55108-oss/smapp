package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWT struct {
	secret []byte
	ttl    time.Duration
}

func NewJWT(secret []byte, ttl time.Duration) *JWT {
	return &JWT{
		secret: secret,
		ttl:    ttl,
	}
}

func (svc *JWT) GenerateJWT(userID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID.String(),
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(svc.ttl).Unix(),
	})
	return token.SignedString(svc.secret)
}

func (svc *JWT) ParseJWT(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return svc.secret, nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if sub, ok := claims["sub"].(string); ok {
			userID, err := uuid.Parse(sub)
			if err != nil {
				return uuid.Nil, fmt.Errorf("parse sub claim: %w", err)
			}
			return userID, nil
		}
		return uuid.Nil, errors.New("expected sub claim to be present and to be a string")
	}
	return uuid.Nil, errors.New("expected claims to be jwt.MapClaims")
}
