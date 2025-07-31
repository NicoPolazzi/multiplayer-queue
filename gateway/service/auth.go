package service

import (
	"errors"
	"time"

	"github.com/NicoPolazzi/multiplayer-queue/gateway/models"
	"github.com/NicoPolazzi/multiplayer-queue/gateway/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsernameTaken      = errors.New("username already taken")
	ErrInvalidCredentials = errors.New("invalid username or password")
)

type AuthService interface {
	Register(username, password string) error
	Login(username, password string) (string, error)
}

type jwtAuthService struct {
	userRepository repository.UserRepository
	key            []byte
}

func NewJWTAuthService(repository repository.UserRepository, key []byte) AuthService {
	return &jwtAuthService{userRepository: repository, key: key}
}

func (s *jwtAuthService) Register(username, password string) error {
	_, err := s.userRepository.FindByUsername(username)

	if err == nil {
		return ErrUsernameTaken
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return s.userRepository.Save(&models.User{Username: username, Password: string(hashedPassword)})
}

func (s *jwtAuthService) Login(username, password string) (string, error) {
	user, err := s.userRepository.FindByUsername(username)

	if errors.Is(err, repository.ErrUserNotFound) {
		return "", ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", ErrInvalidCredentials
	}

	claims := jwt.MapClaims{
		"sub": user.Username,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	result, _ := token.SignedString(s.key)
	return result, nil
}
