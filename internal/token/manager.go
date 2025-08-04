package token

import (
	"errors"
	"time"
)

var (
	ErrImpossibleCreation = errors.New("impossible to create token")
	ErrInvalidToken       = errors.New("invalid token")
)

type TokenManager interface {
	Create(username string, duration time.Duration) (string, error)

	// Validates the token and returns the username
	Validate(token string) (string, error)
}
