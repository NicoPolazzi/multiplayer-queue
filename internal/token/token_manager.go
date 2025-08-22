package token

import (
	"errors"
)

var (
	ErrImpossibleCreation = errors.New("impossible to create token")
	ErrInvalidToken       = errors.New("invalid token")
)

type TokenManager interface {
	Create(username string) (string, error)

	// Validates the token and returns the username
	Validate(token string) (string, error)
}
