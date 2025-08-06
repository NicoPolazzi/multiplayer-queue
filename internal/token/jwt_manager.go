package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// package-level variable used for test purpose only.
var signedString = (*jwt.Token).SignedString

type claims struct {
	Username string `json:"sub"`
	jwt.RegisteredClaims
}

type jwtTokenManager struct {
	secretKey []byte
}

func NewJWTTokenManager(secreteKey []byte) TokenManager {
	return &jwtTokenManager{secretKey: secreteKey}
}

func (j *jwtTokenManager) Create(username string) (string, error) {
	claims := claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := signedString(token, j.secretKey)
	if err != nil {
		return "", ErrImpossibleCreation
	}

	return tokenString, nil
}

func (j *jwtTokenManager) Validate(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (any, error) {
		return j.secretKey, nil
	})

	if err != nil || !token.Valid {
		return "", ErrInvalidToken
	}

	return token.Claims.(*claims).Username, nil
}
