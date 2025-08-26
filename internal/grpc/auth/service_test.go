package auth

import (
	"context"
	"errors"
	"testing"

	pb "github.com/NicoPolazzi/multiplayer-queue/gen/auth"
	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	usrrepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/user"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}
func (m *MockUserRepository) FindByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *MockUserRepository) FindByID(id uint) (*models.User, error) {
	return nil, nil
}

type MockTokenManager struct {
	mock.Mock
}

func (m *MockTokenManager) Create(username string) (string, error) {
	args := m.Called(username)
	return args.String(0), args.Error(1)
}
func (m *MockTokenManager) Validate(token string) (string, error) {
	return "", nil
}

type AuthServerTestSuite struct {
	suite.Suite
	usrRepo    *MockUserRepository
	jwtManager *MockTokenManager
	server     pb.AuthServiceServer
}

func (s *AuthServerTestSuite) SetupTest() {
	s.usrRepo = new(MockUserRepository)
	s.jwtManager = new(MockTokenManager)
	s.server = NewAuthService(s.usrRepo, s.jwtManager)
}

func (s *AuthServerTestSuite) TestRegisterUserSuccess() {
	req := &pb.RegisterUserRequest{Username: "newuser", Password: "password123"}
	s.usrRepo.On("FindByUsername", "newuser").Return(nil, usrrepo.ErrUserNotFound)
	// The call to Run() is necessary because it simulates the behaviour of GORM Create()
	s.usrRepo.On("Create", mock.AnythingOfType("*models.User")).
		Run(func(args mock.Arguments) {
			userArg := args.Get(0).(*models.User)
			userArg.ID = 1
		}).Return(nil)

	resp, err := s.server.RegisterUser(context.Background(), req)
	s.NoError(err)
	s.Equal("newuser", resp.Username)
	s.Equal(uint32(1), resp.Id)
	s.usrRepo.AssertExpectations(s.T())
}

func (s *AuthServerTestSuite) TestRegisterUserWhenUsernameIsTaken() {
	req := &pb.RegisterUserRequest{Username: "existinguser", Password: "password123"}
	mockUser := &models.User{Username: "existinguser"}
	s.usrRepo.On("FindByUsername", "existinguser").Return(mockUser, nil)
	resp, err := s.server.RegisterUser(context.Background(), req)

	s.Empty(resp)
	st, ok := status.FromError(err)
	s.True(ok)
	s.Equal(codes.AlreadyExists, st.Code())
	s.usrRepo.AssertExpectations(s.T())
	s.usrRepo.AssertNotCalled(s.T(), "Create", mock.Anything)
}

func (s *AuthServerTestSuite) TestRegisterUserWhenCreateFails() {
	req := &pb.RegisterUserRequest{Username: "newuser", Password: "password123"}
	dbError := errors.New("database error")
	s.usrRepo.On("FindByUsername", "newuser").Return(nil, usrrepo.ErrUserNotFound)
	s.usrRepo.On("Create", mock.AnythingOfType("*models.User")).Return(dbError)

	resp, err := s.server.RegisterUser(context.Background(), req)

	s.Empty(resp)
	st, ok := status.FromError(err)
	s.True(ok)
	s.Equal(codes.Internal, st.Code())
	s.usrRepo.AssertExpectations(s.T())
}

func (s *AuthServerTestSuite) TestRegisterUserWhenPasswordHashingFails() {
	// bcrypt fails for passwords longer than 72 bytes.
	longPassword := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	req := &pb.RegisterUserRequest{Username: "newuser", Password: longPassword}
	s.usrRepo.On("FindByUsername", "newuser").Return(nil, usrrepo.ErrUserNotFound)

	resp, err := s.server.RegisterUser(context.Background(), req)

	s.Empty(resp)
	st, ok := status.FromError(err)
	s.True(ok)
	s.Equal(codes.Internal, st.Code())
	s.Contains(st.Message(), "failed to hash password")
	s.usrRepo.AssertNotCalled(s.T(), "Create", mock.Anything)
	s.usrRepo.AssertExpectations(s.T())
}

func (s *AuthServerTestSuite) TestLoginUserSuccess() {
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	mockUser := &models.User{Username: "testuser", Password: string(hashedPassword)}
	mockUser.ID = 1
	req := &pb.LoginUserRequest{Username: "testuser", Password: password}
	s.usrRepo.On("FindByUsername", "testuser").Return(mockUser, nil)
	s.jwtManager.On("Create", "testuser").Return("mock-jwt-token", nil)

	resp, err := s.server.LoginUser(context.Background(), req)

	s.NoError(err)
	s.NotNil(resp)
	s.Equal("mock-jwt-token", resp.Token)
	s.Equal(uint32(1), resp.User.Id)
	s.Equal("testuser", resp.User.Username)
	s.usrRepo.AssertExpectations(s.T())
	s.jwtManager.AssertExpectations(s.T())
}

func (s *AuthServerTestSuite) TestLoginUserWhenUserNotFound() {
	req := &pb.LoginUserRequest{Username: "unknown", Password: "password123"}
	s.usrRepo.On("FindByUsername", "unknown").Return(nil, usrrepo.ErrUserNotFound)

	resp, err := s.server.LoginUser(context.Background(), req)

	s.Empty(resp)
	st, ok := status.FromError(err)
	s.True(ok)
	s.Equal(codes.NotFound, st.Code())
	s.Equal(st.Message(), "invalid credentials")
	s.jwtManager.AssertNotCalled(s.T(), "Create", mock.Anything)
}

func (s *AuthServerTestSuite) TestLoginUserWithWrongPassword() {
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	mockUser := &models.User{Username: "testuser", Password: string(hashedPassword)}
	req := &pb.LoginUserRequest{Username: "testuser", Password: "wrongpassword"}
	s.usrRepo.On("FindByUsername", "testuser").Return(mockUser, nil)

	resp, err := s.server.LoginUser(context.Background(), req)

	s.Empty(resp)
	st, ok := status.FromError(err)
	s.True(ok)
	s.Equal(codes.Unauthenticated, st.Code())
	s.jwtManager.AssertNotCalled(s.T(), "Create", mock.Anything)
}

func (s *AuthServerTestSuite) TestLoginUserWhenTokenCreationFails() {
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	mockUser := &models.User{Username: "testuser", Password: string(hashedPassword)}
	req := &pb.LoginUserRequest{Username: "testuser", Password: password}
	tokenError := errors.New("jwt error")
	s.usrRepo.On("FindByUsername", "testuser").Return(mockUser, nil)
	s.jwtManager.On("Create", "testuser").Return("", tokenError)

	resp, err := s.server.LoginUser(context.Background(), req)

	s.Empty(resp)
	st, ok := status.FromError(err)
	s.True(ok)
	s.Equal(codes.Internal, st.Code())
	s.jwtManager.AssertExpectations(s.T())
}

func (s *AuthServerTestSuite) TestLoginUserWhenCanNotRetrieveUser() {
	req := &pb.LoginUserRequest{Username: "existing", Password: "password123"}
	s.usrRepo.On("FindByUsername", "existing").Return(nil, errors.New("internal error"))

	resp, err := s.server.LoginUser(context.Background(), req)

	s.Empty(resp)
	st, ok := status.FromError(err)
	s.True(ok)
	s.Equal(codes.Internal, st.Code())
	s.Equal(st.Message(), "failed to retrieve user: internal error")
	s.jwtManager.AssertNotCalled(s.T(), "Create", mock.Anything)
}

func TestAuthServer(t *testing.T) {
	suite.Run(t, new(AuthServerTestSuite))
}
