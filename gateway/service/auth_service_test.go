package service

import (
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/NicoPolazzi/multiplayer-queue/gateway/models"
	"github.com/NicoPolazzi/multiplayer-queue/gateway/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AuthServiceTestSuite struct {
	suite.Suite
}

// This is the strandard behaviour of creating a mock object with testify
type UserTestRepository struct {
	mock.Mock
}

// These calls work with setting up an On and Return before actually the real methods are called
// in the program
func (r *UserTestRepository) Save(user *models.User) error {
	// Tells the mock object that the method has been called.
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

func (s *AuthServiceTestSuite) TestRegisterSuccess() {
	username := "test"
	password := "123"

	testRepository := new(UserTestRepository)

	mock.InOrder(
		testRepository.On("FindByUsername", username).Return(nil, repository.ErrUserNotFound),
		testRepository.On("Save", mock.MatchedBy(func(user *models.User) bool {
			return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil
		})).Return(nil),
	)

	authService := NewAuthService(testRepository)
	err := authService.Register(username, password)

	assert.Nil(s.T(), err)
	testRepository.AssertExpectations(s.T())
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}
