package service

import (
	"errors"

	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	repository "github.com/NicoPolazzi/multiplayer-queue/internal/repository/user"
	"github.com/NicoPolazzi/multiplayer-queue/internal/token"
	"golang.org/x/crypto/bcrypt"
)

// package-level variable used for test purpose only. This is necessary because I don't want to mock the hasher.
var bcryptGenerate = bcrypt.GenerateFromPassword

type JWTAuthService struct {
	userRepository repository.UserRepository
	jwtManager     token.TokenManager
}

func NewJWTAuthService(repository repository.UserRepository) AuthService {
	return &JWTAuthService{userRepository: repository}
}

func (s *JWTAuthService) SetTokenManager(m token.TokenManager) {
	s.jwtManager = m
}

func (s *JWTAuthService) Register(username, password string) error {
	if _, err := s.userRepository.FindByUsername(username); err == nil {
		return ErrUsernameTaken
	}

	hashedPassword, err := bcryptGenerate([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepository.Create(&models.User{Username: username, Password: string(hashedPassword)})
}

func (s *JWTAuthService) Login(username, password string) (string, error) {
	user, err := s.userRepository.FindByUsername(username)
	if errors.Is(err, repository.ErrUserNotFound) {
		return "", ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", ErrInvalidCredentials
	}

	return s.jwtManager.Create(username)
}
