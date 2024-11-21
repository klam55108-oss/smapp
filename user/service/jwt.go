package service

import (
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWT struct {
	privateKey *rsa.PrivateKey
	ttl        time.Duration
}

func NewJWT(privateKey *rsa.PrivateKey, ttl time.Duration) *JWT {
	return &JWT{
		privateKey: privateKey,
		ttl:        ttl,
	}
}

func (svc *JWT) GenerateJWT(userID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": userID.String(),
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(svc.ttl).Unix(),
	})
	return token.SignedString(svc.privateKey)
}
