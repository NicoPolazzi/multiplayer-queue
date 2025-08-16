package service

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	usrrepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	UserFixtureUsername string = "test"
	UserFixturePassword string = "123"
)

type JWTAuthServiceTestSuite struct {
	suite.Suite
	Repository  *UserTestRepository
	Manager     *TokenTestManager
	authService AuthService
}

type TokenTestManager struct {
	mock.Mock
}

func (t *TokenTestManager) Create(username string) (string, error) {
	args := t.Called(username)
	return args.String(0), args.Error(1)
}

func (t *TokenTestManager) Validate(token string) (string, error) {
	args := t.Called(token)
	return args.String(0), args.Error(1)
}

type UserTestRepository struct {
	mock.Mock
}

func (r *UserTestRepository) Create(user *models.User) error {
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

func (r *UserTestRepository) FindByID(id uint) (*models.User, error) {
	args := r.Called(id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (s *JWTAuthServiceTestSuite) SetupTest() {
	s.Repository = new(UserTestRepository)
	s.Manager = new(TokenTestManager)
	s.authService = NewJWTAuthService(s.Repository, s.Manager)
}

func (s *JWTAuthServiceTestSuite) TestRegisterWhenThereIsNotARegisteredUserShouldSuccess() {
	mock.InOrder(
		s.Repository.On("FindByUsername", UserFixtureUsername).Return(nil, usrrepo.ErrUserNotFound),
		s.Repository.On("Create", mock.MatchedBy(func(user *models.User) bool {
			return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(UserFixturePassword)) == nil
		})).Return(nil),
	)

	err := s.authService.Register(UserFixtureUsername, UserFixturePassword)
	assert.Nil(s.T(), err)
	s.Repository.AssertExpectations(s.T())
}

func (s *JWTAuthServiceTestSuite) TestRegisterWhenThereIsAlreadyAnUserShouldRaiseUsernameTakenError() {
	s.Repository.On("FindByUsername", UserFixtureUsername).Return(models.User{}, nil)
	err := s.authService.Register(UserFixtureUsername, UserFixturePassword)
	s.Repository.AssertExpectations(s.T())
	assert.ErrorIs(s.T(), err, ErrUsernameTaken)
}

func (s *JWTAuthServiceTestSuite) TestRegisterOnHashErrorShouldFail() {
	// This is necessary to let other tests to use the regular function
	originalBcryptGenerate := bcryptGenerate
	defer func() { bcryptGenerate = originalBcryptGenerate }()
	bcryptGenerate = func(password []byte, cost int) ([]byte, error) {
		return nil, errors.New("mock hash failure")
	}

	s.Repository.On("FindByUsername", UserFixtureUsername).Return(nil, usrrepo.ErrUserNotFound)
	err := s.authService.Register(UserFixtureUsername, UserFixturePassword)
	s.Repository.AssertExpectations(s.T())
	assert.Error(s.T(), err)
}

func (s *JWTAuthServiceTestSuite) TestLoginSuccess() {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(UserFixturePassword), bcrypt.DefaultCost)
	s.Repository.On("FindByUsername", UserFixtureUsername).Return(
		&models.User{Username: UserFixtureUsername, Password: string(hashedPassword)}, nil)
	s.Manager.On("Create", UserFixtureUsername).Return("mock-jwt-token-value", nil)

	token, err := s.authService.Login(UserFixtureUsername, UserFixturePassword)

	s.Repository.AssertExpectations(s.T())
	s.Manager.AssertExpectations(s.T())
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "mock-jwt-token-value", token)
}

func (s *JWTAuthServiceTestSuite) TestLoginWhenUserIsNotFoundShouldReturnInvalidCredentialsError() {
	s.Repository.On("FindByUsername", UserFixtureUsername).Return(nil, usrrepo.ErrUserNotFound)
	token, err := s.authService.Login(UserFixtureUsername, UserFixturePassword)
	s.Repository.AssertExpectations(s.T())
	assert.ErrorIs(s.T(), err, ErrInvalidCredentials)
	assert.Empty(s.T(), token)
}

func (s *JWTAuthServiceTestSuite) TestLoginWhenPasswordDoesNotMatchdReturnInvalidCredentialsError() {
	user := models.User{Username: UserFixtureUsername, Password: UserFixturePassword}
	s.Repository.On("FindByUsername", UserFixtureUsername).Return(&user, nil)
	token, err := s.authService.Login(UserFixtureUsername, "wrong password")
	assert.ErrorIs(s.T(), err, ErrInvalidCredentials)
	assert.Empty(s.T(), token)
}

func TestJWTAuthService(t *testing.T) {
	suite.Run(t, new(JWTAuthServiceTestSuite))
}
