package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type MiddlewareUtilsTestSuite struct {
	suite.Suite
	recorder *httptest.ResponseRecorder
	context  *gin.Context
}

func (s *MiddlewareUtilsTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.recorder = httptest.NewRecorder()
	s.context, _ = gin.CreateTestContext(s.recorder)
}

func (s *MiddlewareUtilsTestSuite) TestEnsureLoggedInWhenUserIsLoggedIn() {
	s.context.Set("is_logged_in", true)
	s.context.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	handler := EnsureLoggedIn()
	handler(s.context)
	s.NotEqual(http.StatusSeeOther, s.recorder.Code, "Should not redirect when logged in")
}

func (s *MiddlewareUtilsTestSuite) TestEnsureLoggedInWhenUserIsNotLoggedIn() {
	s.context.Set("is_logged_in", false)
	s.context.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	handler := EnsureLoggedIn()
	handler(s.context)
	s.True(s.context.IsAborted())
	s.Equal(http.StatusSeeOther, s.recorder.Code, "Should redirect when not logged in")
	s.Equal("/user/login", s.recorder.Header().Get("Location"), "Should redirect to login page")
}

func (s *MiddlewareUtilsTestSuite) TestEnsureLoggedInWhenIsLoggedInNotSet() {
	s.context.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	handler := EnsureLoggedIn()
	handler(s.context)
	s.True(s.context.IsAborted())
	s.Equal(http.StatusSeeOther, s.recorder.Code, "Should redirect when not logged in")
	s.Equal(handlers.LoginPagePath, s.recorder.Header().Get("Location"), "Should redirect to login page")
}

func (s *MiddlewareUtilsTestSuite) TestEnsureNotLoggedInWhenUserIsNotLoggedIn() {
	s.context.Set("is_logged_in", false)
	s.context.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	handler := EnsureNotLoggedIn()
	handler(s.context)
	s.NotEqual(http.StatusSeeOther, s.recorder.Code, "Should not redirect when not logged in")
}

func (s *MiddlewareUtilsTestSuite) TestEnsureNotLoggedInWhenUserIsLoggedIn() {
	s.context.Set("is_logged_in", true)
	s.context.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	handler := EnsureNotLoggedIn()
	handler(s.context)
	s.True(s.context.IsAborted())
	s.Equal(http.StatusSeeOther, s.recorder.Code, "Should redirect when logged in")
	s.Equal(handlers.IndexPagePath, s.recorder.Header().Get("Location"), "Should redirect to home page")
}

func TestMiddlewareUtils(t *testing.T) {
	suite.Run(t, new(MiddlewareUtilsTestSuite))
}
