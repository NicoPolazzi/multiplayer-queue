package gorm

import (
	"errors"

	"github.com/NicoPolazzi/multiplayer-queue/gateway/models"
	"github.com/NicoPolazzi/multiplayer-queue/gateway/repository"
	"gorm.io/gorm"
)

type GormUserRepository struct {
	DB *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) repository.UserRepository {
	return &GormUserRepository{
		DB: db,
	}
}

func (r *GormUserRepository) Save(user *models.User) error {
	if result := r.DB.Save(user); result.Error != nil {
		return repository.ErrUserExists
	} else {
		return nil
	}

}

func (r *GormUserRepository) FindByUsername(username string) (*models.User, error) {
	var retrievedUser models.User
	result := r.DB.Where(&models.User{Username: username}).First(&retrievedUser)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, repository.ErrUserNotFound
	}

	return &retrievedUser, nil
}
