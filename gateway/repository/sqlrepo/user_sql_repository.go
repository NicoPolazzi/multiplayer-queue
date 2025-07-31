package sqlrepo

import (
	"errors"

	"github.com/NicoPolazzi/multiplayer-queue/gateway/models"
	"github.com/NicoPolazzi/multiplayer-queue/gateway/repository"
	"gorm.io/gorm"
)

type sqlUserRepository struct {
	DB *gorm.DB
}

func NewSQLUserRepository(db *gorm.DB) repository.UserRepository {
	return &sqlUserRepository{
		DB: db,
	}
}

func (r *sqlUserRepository) Save(user *models.User) error {
	if result := r.DB.Save(user); result.Error != nil {
		return repository.ErrUserExists
	} else {
		return nil
	}
}

func (r *sqlUserRepository) FindByUsername(username string) (*models.User, error) {
	var retrievedUser models.User
	result := r.DB.Where(&models.User{Username: username}).First(&retrievedUser)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, repository.ErrUserNotFound
	}

	return &retrievedUser, nil
}
