package usrrepo

import (
	"errors"

	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	"gorm.io/gorm"
)

type sqlUserRepository struct {
	DB *gorm.DB
}

func NewSQLUserRepository(db *gorm.DB) UserRepository {
	return &sqlUserRepository{
		DB: db,
	}
}

func (r *sqlUserRepository) Save(user *models.User) error {
	if result := r.DB.Save(user); result.Error != nil {
		return ErrUserExists
	} else {
		return nil
	}
}

func (r *sqlUserRepository) FindByUsername(username string) (*models.User, error) {
	var retrievedUser models.User
	result := r.DB.Where(&models.User{Username: username}).First(&retrievedUser)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}

	return &retrievedUser, nil
}
