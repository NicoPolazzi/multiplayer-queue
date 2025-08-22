package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/internal/token"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockTokenManager struct {
	mock.Mock
}

func (m *MockTokenManager) Create(username string) (string, error) {
	return "", nil
}

func (m *MockTokenManager) Validate(tokenString string) (string, error) {
	args := m.Called(tokenString)
	return args.String(0), args.Error(1)
}

type AuthMiddlewareTestSuite struct {
	suite.Suite
	tokenManager   *MockTokenManager
	authMiddleware *AuthMiddleware
	recorder       *httptest.ResponseRecorder
	context        *gin.Context
}

func (s *AuthMiddlewareTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.tokenManager = new(MockTokenManager)
	s.authMiddleware = NewAuthMiddleware(s.tokenManager)
	s.recorder = httptest.NewRecorder()
	s.context, _ = gin.CreateTestContext(s.recorder)
}

func (s *AuthMiddlewareTestSuite) TestAuthMiddlewareWhenTokenIsValid() {
	validToken := "valid.token.string"
	expectedUsername := "testuser"
	s.tokenManager.On("Validate", validToken).Return(expectedUsername, nil)

	s.context.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	s.context.Request.AddCookie(&http.Cookie{
		Name:  "token",
		Value: validToken,
	})

	handler := s.authMiddleware.CheckUser()
	handler(s.context)

	s.tokenManager.AssertExpectations(s.T())
	username, _ := s.context.Get("username")
	isLoggedIn, _ := s.context.Get("is_logged_in")
	s.True(isLoggedIn.(bool))
	s.Equal(expectedUsername, username)
	s.False(s.context.IsAborted())
}

func (s *AuthMiddlewareTestSuite) TestAuthMiddlewareWhenCookiesIsNotSet() {
	s.context.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	handler := s.authMiddleware.CheckUser()
	handler(s.context)
	username, _ := s.context.Get("username")
	isLoggedIn, _ := s.context.Get("is_logged_in")
	s.Empty(username)
	s.Empty(isLoggedIn)
}

func (s *AuthMiddlewareTestSuite) TestAuthMiddlewareWhenTokenIsInvalid() {
	invalidToken := "invalid.token.string"
	s.tokenManager.On("Validate", invalidToken).Return("", token.ErrInvalidToken)
	s.context.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	s.context.Request.AddCookie(&http.Cookie{
		Name:  "token",
		Value: invalidToken,
	})

	handler := s.authMiddleware.CheckUser()
	handler(s.context)
	s.tokenManager.AssertExpectations(s.T())
	username, _ := s.context.Get("username")
	isLoggedIn, _ := s.context.Get("is_logged_in")
	s.Empty(username)
	s.False(isLoggedIn.(bool))
}

func TestAuthMiddleware(t *testing.T) {
	suite.Run(t, new(AuthMiddlewareTestSuite))
}
