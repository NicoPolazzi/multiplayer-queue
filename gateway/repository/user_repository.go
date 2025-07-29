package repository

import (
	"errors"

	"github.com/NicoPolazzi/multiplayer-queue/gateway/models"
)

var (
	ErrUserExists   = errors.New("user already exists in the database")
	ErrUserNotFound = errors.New("user not found in the database")
)

type UserRepository interface {
	Save(user *models.User) error
	FindByUsername(username string) (*models.User, error)
}
