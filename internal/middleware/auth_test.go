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
	args := m.Called(username)
	return args.String(0), args.Error(1)
}

func (m *MockTokenManager) Validate(tokenString string) (string, error) {
	args := m.Called(tokenString)
	return args.String(0), args.Error(1)
}

type AuthMiddlewareTestSuite struct {
	suite.Suite
	tokenManager *MockTokenManager
	recorder     *httptest.ResponseRecorder
	context      *gin.Context
}

func (s *AuthMiddlewareTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.tokenManager = new(MockTokenManager)
	s.recorder = httptest.NewRecorder()
	s.context, _ = gin.CreateTestContext(s.recorder)
}

func (s *AuthMiddlewareTestSuite) TestAuthMiddlewareWhenTokenIsValidSucceeds() {
	validToken := "valid.token.string"
	expectedUsername := "testuser"
	s.tokenManager.On("Validate", validToken).Return(expectedUsername, nil)

	s.context.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	s.context.Request.AddCookie(&http.Cookie{
		Name:  "jwt",
		Value: validToken,
	})

	handler := AuthMiddleware(s.tokenManager)
	handler(s.context)

	s.tokenManager.AssertExpectations(s.T())
	s.False(s.context.IsAborted())
	s.Equal(expectedUsername, s.context.GetString("username"))
}

func (s *AuthMiddlewareTestSuite) TestAuthMiddlewareWhenCookieIsMissingAbortAndRedirectToLogin() {
	s.context.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	handler := AuthMiddleware(s.tokenManager)
	handler(s.context)
	s.True(s.context.IsAborted())
	s.Equal(http.StatusSeeOther, s.recorder.Code)
	s.Equal("/login", s.recorder.Header().Get("Location"))
}

func (s *AuthMiddlewareTestSuite) TestAuthMiddlewareWhenRequestIsInvalidAbortContext() {
	invalidToken := "invalidToken"
	s.tokenManager.On("Validate", invalidToken).Return("", token.ErrInvalidToken)
	s.context.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	s.context.Request.AddCookie(&http.Cookie{
		Name:  "jwt",
		Value: invalidToken,
	})

	handler := AuthMiddleware(s.tokenManager)
	handler(s.context)

	s.tokenManager.AssertExpectations(s.T())
	s.True(s.context.IsAborted())
	s.Equal(http.StatusUnauthorized, s.recorder.Code)
	s.JSONEq(`{"status": "error", "message": "invalid token"}`,
		s.recorder.Body.String())
}

func TestAuthMiddleware(t *testing.T) {
	suite.Run(t, new(AuthMiddlewareTestSuite))
}
