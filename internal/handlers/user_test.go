package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(username, password string) error {
	args := m.Called(username, password)
	return args.Error(0)
}

func (m *MockAuthService) Login(username, password string) (string, error) {
	args := m.Called(username, password)
	return args.String(0), args.Error(1)
}

type UserHandlerTestSuite struct {
	suite.Suite
	authService *MockAuthService
	handler     *UserHandler
	recorder    *httptest.ResponseRecorder
	context     *gin.Context
	engine      *gin.Engine
}

func (s *UserHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.authService = new(MockAuthService)
	s.handler = NewUserHandler(s.authService)
	s.recorder = httptest.NewRecorder()
	s.context, s.engine = gin.CreateTestContext(s.recorder)
	s.engine.LoadHTMLGlob("../../web/templates/*")
}

func (s *UserHandlerTestSuite) TestShowIndexPage() {
	s.context.Set("is_logged_in", true)
	s.context.Set("username", "testuser")
	s.handler.ShowIndexPage(s.context)
	s.Equal(http.StatusOK, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), "Home Page")
	s.Contains(s.recorder.Body.String(), "testuser")
}

func (s *UserHandlerTestSuite) TestShowRegisterPage() {
	s.handler.ShowLRegisterPage(s.context)
	s.Equal(http.StatusOK, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), "<title>Register</title>")
}

func (s *UserHandlerTestSuite) TestPerformRegistrationWhenRequestIsMalformed() {
	s.setPostRequest("")
	s.handler.PerformRegistration(s.context)
	s.Equal(http.StatusBadRequest, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), RegistrationErrorMessage)
}
func (s *UserHandlerTestSuite) setPostRequest(form string) {
	s.context.Request = httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(form))
	s.context.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
}

func (s *UserHandlerTestSuite) TestPerformRegistrationWhenUsernameTaken() {
	s.authService.On("Register", "takenuser", "pass").Return(service.ErrUsernameTaken)
	s.setPostRequest("username=takenuser&password=pass")
	s.handler.PerformRegistration(s.context)
	s.Equal(http.StatusBadRequest, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), RegistrationErrorMessage)
	s.authService.AssertExpectations(s.T())
}

func (s *UserHandlerTestSuite) TestPerformRegistrationWhenInternalError() {
	s.authService.On("Register", "takenuser", "pass").Return(errors.New("some internal error"))
	s.setPostRequest("username=takenuser&password=pass")
	s.handler.PerformRegistration(s.context)
	s.Equal(http.StatusInternalServerError, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), RegistrationErrorMessage)
	s.authService.AssertExpectations(s.T())
}

func (s *UserHandlerTestSuite) TestPerformRegistrationSuccess() {
	s.authService.On("Register", "newuser", "pass").Return(nil)
	s.setPostRequest("username=newuser&password=pass")
	s.handler.PerformRegistration(s.context)
	s.Equal(http.StatusOK, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), "Successful registration")
	s.authService.AssertExpectations(s.T())
}

func (s *UserHandlerTestSuite) TestShowLoginPage() {
	s.handler.ShowLoginPage(s.context)
	s.Equal(http.StatusOK, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), "<title>Login</title>")
}

func (s *UserHandlerTestSuite) TestPerformLoginWhenRequestIsMalformed() {
	s.setPostRequest("")
	s.handler.PerformLogin(s.context)
	s.Equal(http.StatusBadRequest, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), LoginErrorMessage)
}

func (s *UserHandlerTestSuite) TestPerformLoginWhenInvalidCredentials() {
	s.authService.On("Login", "testuser", "invalid").Return("", service.ErrInvalidCredentials)
	s.setPostRequest("username=testuser&password=invalid")
	s.handler.PerformLogin(s.context)
	s.Equal(http.StatusUnauthorized, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), LoginErrorMessage)
	s.authService.AssertExpectations(s.T())
}

func (s *UserHandlerTestSuite) TestPerformLoginWhenInternalError() {
	s.authService.On("Login", "testuser", "invalid").Return("", errors.New("some internal error"))
	s.setPostRequest("username=testuser&password=invalid")
	s.handler.PerformLogin(s.context)
	s.Equal(http.StatusInternalServerError, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), LoginErrorMessage)
	s.authService.AssertExpectations(s.T())
}

func (s *UserHandlerTestSuite) TestPerformLoginSuccess() {
	token := "sometoken"
	s.authService.On("Login", "testuser", "goodpassword").Return(token, nil)
	s.setPostRequest("username=testuser&password=goodpassword")
	s.handler.PerformLogin(s.context)
	s.Equal(http.StatusOK, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), "<title>Successful Login</title>")
	s.Equal(s.getJWTCookieValue(), token, "JWT cookie should be set with the token")
	s.authService.AssertExpectations(s.T())
}

func (s *UserHandlerTestSuite) getJWTCookieValue() string {
	for _, cookie := range s.recorder.Result().Cookies() {
		if cookie.Name == "jwt" {
			return cookie.Value
		}
	}
	return ""
}

func (s *UserHandlerTestSuite) TestPerformLogout() {
	s.context.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	s.handler.PerformLogout(s.context)
	s.Equal(s.getJWTCookieValue(), "", "JWT cookie should be cleared on logout")
	s.Equal(http.StatusSeeOther, s.recorder.Code)
	s.Equal(IndexPagePath, s.recorder.Header().Get("Location"))
}

func TestUserHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}
