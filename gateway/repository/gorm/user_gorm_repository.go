package gorm

import (
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
	return &models.User{}, nil
}
