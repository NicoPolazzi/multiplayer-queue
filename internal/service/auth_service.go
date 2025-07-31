package service

import "errors"

var (
	ErrUsernameTaken      = errors.New("username already taken")
	ErrInvalidCredentials = errors.New("invalid username or password")
)

type AuthService interface {
	Register(username, password string) error
	Login(username, password string) (string, error)
}
