package usrrepo

import (
	"errors"

	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	"gorm.io/gorm"
)

type sqlUserRepository struct {
	db *gorm.DB
}

func NewSQLUserRepository(db *gorm.DB) UserRepository {
	return &sqlUserRepository{
		db: db,
	}
}

func (r *sqlUserRepository) Create(user *models.User) error {
	if result := r.db.Create(user); result.Error != nil {
		return ErrUserExists
	} else {
		return nil
	}
}

func (r *sqlUserRepository) FindByUsername(username string) (*models.User, error) {
	var retrievedUser models.User
	result := r.db.Where(&models.User{Username: username}).First(&retrievedUser)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}

	return &retrievedUser, nil
}

func (r *sqlUserRepository) FindByID(id uint) (*models.User, error) {
	var retrievedUser models.User
	result := r.db.First(&retrievedUser, id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}

	return &retrievedUser, nil
}
