package service

import (
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/NicoPolazzi/multiplayer-queue/gateway/models"
	"github.com/NicoPolazzi/multiplayer-queue/gateway/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	UserFixtureUsername string = "test"
	UserFixturePassword string = "123"
)

type AuthServiceTestSuite struct {
	suite.Suite
	Repository *UserTestRepository
	AuthService
}

type UserTestRepository struct {
	mock.Mock
}

func (r *UserTestRepository) Save(user *models.User) error {
	args := r.Called(user)
	return args.Error(0)
}

func (r *UserTestRepository) FindByUsername(username string) (*models.User, error) {
	args := r.Called(username)
	if user, ok := args.Get(0).(*models.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (s *AuthServiceTestSuite) SetupTest() {
	s.Repository = new(UserTestRepository)
	s.AuthService = NewJWTAuthService(s.Repository, []byte("test-secret"))
}

func (s *AuthServiceTestSuite) TestRegisterWhenThereIsNotARegisteredUserShouldSuccess() {
	mock.InOrder(
		s.Repository.On("FindByUsername", UserFixtureUsername).Return(nil, repository.ErrUserNotFound),
		s.Repository.On("Save", mock.MatchedBy(func(user *models.User) bool {
			return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(UserFixturePassword)) == nil
		})).Return(nil),
	)

	err := s.AuthService.Register(UserFixtureUsername, UserFixturePassword)
	assert.Nil(s.T(), err)
	s.Repository.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestRegisterWhenThereIsAlreadyAnUserShouldRaiseAnError() {
	s.Repository.On("FindByUsername", UserFixtureUsername).Return(mock.AnythingOfType("*models.User"), nil)
	err := s.AuthService.Register(UserFixtureUsername, UserFixturePassword)
	s.Repository.AssertExpectations(s.T())
	assert.ErrorIs(s.T(), err, ErrUsernameTaken)
}

func (s *AuthServiceTestSuite) TestLoginSuccess() {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(UserFixturePassword), bcrypt.DefaultCost)
	s.Repository.On("FindByUsername", UserFixtureUsername).Return(
		&models.User{Username: UserFixtureUsername, Password: string(hashedPassword)}, nil)

	token, err := s.AuthService.Login(UserFixtureUsername, UserFixturePassword)

	claims := jwt.MapClaims{
		"sub": UserFixtureUsername,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expected, _ := t.SignedString([]byte("test-secret"))

	s.Repository.AssertExpectations(s.T())
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), expected, token)
}

func (s *AuthServiceTestSuite) TestLoginWhenUserIsNotFoundShouldReturnInvalidCredentialsError() {
	s.Repository.On("FindByUsername", UserFixtureUsername).Return(nil, repository.ErrUserNotFound)
	token, err := s.AuthService.Login(UserFixtureUsername, UserFixturePassword)
	s.Repository.AssertExpectations(s.T())
	assert.ErrorIs(s.T(), err, ErrInvalidCredentials)
	assert.Empty(s.T(), token)
}

func (s *AuthServiceTestSuite) TestLoginWhenPasswordDoesNotMatchdReturnInvalidCredentialsError() {
	user := models.User{Username: UserFixtureUsername, Password: UserFixturePassword}
	s.Repository.On("FindByUsername", UserFixtureUsername).Return(&user, nil)
	token, err := s.AuthService.Login(UserFixtureUsername, "wrong password")
	assert.ErrorIs(s.T(), err, ErrInvalidCredentials)
	assert.Empty(s.T(), token)
}

func TestJWTAuthService(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}
