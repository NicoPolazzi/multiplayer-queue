package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// package-level variable used for test purpose only.
var signedString = (*jwt.Token).SignedString

type JWTTokenManager struct {
	secretKey []byte
}

func NewJWTTokenManager(secreteKey []byte) TokenManager {
	return &JWTTokenManager{secretKey: secreteKey}
}

func (j *JWTTokenManager) Create(username string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := signedString(token, j.secretKey)
	if err != nil {
		return "", ErrImpossibleCreation
	}

	return tokenString, nil
}

func (j *JWTTokenManager) Validate(token string) (string, error) {
	panic("unimplemented")
}
