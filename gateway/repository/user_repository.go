package repository

import "github.com/NicoPolazzi/multiplayer-queue/gateway/models"

type UserRepository interface {
	Create(user *models.User) error
	FindByUsername(username string) (*models.User, error)
}
