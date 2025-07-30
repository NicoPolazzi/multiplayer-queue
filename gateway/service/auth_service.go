package service

import (
	"errors"

	"github.com/NicoPolazzi/multiplayer-queue/gateway/models"
	"github.com/NicoPolazzi/multiplayer-queue/gateway/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsernameTaken = errors.New("username already taken")
)

type AuthService struct {
	UserRepository repository.UserRepository
}

func NewAuthService(repository repository.UserRepository) *AuthService {
	return &AuthService{UserRepository: repository}
}

func (s *AuthService) Register(username, password string) error {
	_, err := s.UserRepository.FindByUsername(username)

	if err == nil {
		return ErrUsernameTaken
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return s.UserRepository.Save(&models.User{Username: username, Password: string(hashedPassword)})
}
