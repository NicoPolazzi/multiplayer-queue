package service

import (
	"errors"

	"github.com/NicoPolazzi/multiplayer-queue/gateway/models"
	"github.com/NicoPolazzi/multiplayer-queue/gateway/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepository repository.UserRepository
}

func NewAuthService(repository repository.UserRepository) *AuthService {
	return &AuthService{UserRepository: repository}
}

func (s *AuthService) Register(username, password string) error {

	if _, err := s.UserRepository.FindByUsername(username); errors.Is(err, repository.ErrUserNotFound) {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		s.UserRepository.Save(&models.User{Username: username, Password: string(hashedPassword)})
	}
	return nil
}
