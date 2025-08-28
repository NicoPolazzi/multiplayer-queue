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
	tokenManager   *MockTokenManager
	authMiddleware *AuthMiddleware
}

func (s *AuthMiddlewareTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.tokenManager = new(MockTokenManager)
	s.authMiddleware = NewAuthMiddleware(s.tokenManager)
}

// Helper to create a test context and recorder
func (s *AuthMiddlewareTestSuite) createTestContext(req *http.Request) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	return w, ctx
}

func (s *AuthMiddlewareTestSuite) TestCheckUserWhenTokenIsValid() {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "valid-token"})
	_, ctx := s.createTestContext(req)

	s.tokenManager.On("Validate", "valid-token").Return("testuser", nil)
	handler := s.authMiddleware.CheckUser()

	handler(ctx)

	user, ok := UserFromContext(ctx)
	s.True(ok, "User should be found in context")
	s.NotNil(user)
	s.Equal("testuser", user.Username)
	s.False(ctx.IsAborted())
	s.tokenManager.AssertExpectations(s.T())
}

func (s *AuthMiddlewareTestSuite) TestCheckUserWhenCookieIsNotSet() {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	_, ctx := s.createTestContext(req)
	handler := s.authMiddleware.CheckUser()

	handler(ctx)
	_, ok := UserFromContext(ctx)
	s.False(ok, "User should not be found in context")
	s.False(ctx.IsAborted())
}

func (s *AuthMiddlewareTestSuite) TestCheckUserWhenTokenIsInvalid() {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "invalid-token"})
	_, ctx := s.createTestContext(req)

	s.tokenManager.On("Validate", "invalid-token").Return("", token.ErrInvalidToken)
	handler := s.authMiddleware.CheckUser()

	handler(ctx)

	_, ok := UserFromContext(ctx)
	s.False(ok, "User should not be found in context")
	s.False(ctx.IsAborted())
	s.tokenManager.AssertExpectations(s.T())
}

func (s *AuthMiddlewareTestSuite) TestEnsureLoggedInSuccess() {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	w, ctx := s.createTestContext(req)
	SetUserInContext(ctx, &User{Username: "testuser"})
	handler := EnsureLoggedIn()

	handler(ctx)

	s.False(ctx.IsAborted())
	s.Equal(http.StatusOK, w.Code, "Expected no redirect")
}

func (s *AuthMiddlewareTestSuite) TestEnsureLoggedInRedirectsWhenNotLoggedIn() {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	w, ctx := s.createTestContext(req)
	handler := EnsureLoggedIn()

	handler(ctx)

	s.True(ctx.IsAborted())
	s.Equal(http.StatusSeeOther, w.Code)
	s.Equal("/user/login", w.Header().Get("Location"))
}

func (s *AuthMiddlewareTestSuite) TestEnsureNotLoggedInSuccess() {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	w, ctx := s.createTestContext(req)
	handler := EnsureNotLoggedIn()

	handler(ctx)
	s.False(ctx.IsAborted())
	s.Equal(http.StatusOK, w.Code)
}

func (s *AuthMiddlewareTestSuite) TestEnsureNotLoggedInRedirectsWhenLoggedIn() {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	w, ctx := s.createTestContext(req)
	SetUserInContext(ctx, &User{Username: "testuser"})
	handler := EnsureNotLoggedIn()

	handler(ctx)

	s.True(ctx.IsAborted())
	s.Equal(http.StatusSeeOther, w.Code)
	s.Equal("/", w.Header().Get("Location"))
}

func TestAuthMiddleware(t *testing.T) {
	suite.Run(t, new(AuthMiddlewareTestSuite))
}
