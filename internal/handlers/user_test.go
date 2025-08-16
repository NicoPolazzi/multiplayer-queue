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
	s.Contains(s.recorder.Body.String(), "Login")
}

func (s *UserHandlerTestSuite) TestPerformRegistrationWhenRequestIsMalformed() {
	s.context.Request = httptest.NewRequest(http.MethodPost, "/test", nil)
	s.context.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	s.handler.PerformRegistration(s.context)
	s.Equal(http.StatusBadRequest, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), RegistrationErrorMessage)
}

func (s *UserHandlerTestSuite) TestPerformRegistrationWhenUsernameTaken() {
	s.authService.On("Register", "takenuser", "pass").Return(service.ErrUsernameTaken)
	form := "username=takenuser&password=pass"
	s.context.Request = httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(form))
	s.context.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	s.handler.PerformRegistration(s.context)
	s.Equal(http.StatusBadRequest, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), RegistrationErrorMessage)
	s.authService.AssertExpectations(s.T())
}

func (s *UserHandlerTestSuite) TestPerformRegistrationWhenInternalError() {
	s.authService.On("Register", "takenuser", "pass").Return(errors.New("some internal error"))
	form := "username=takenuser&password=pass"
	s.context.Request = httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(form))
	s.context.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	s.handler.PerformRegistration(s.context)
	s.Equal(http.StatusInternalServerError, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), RegistrationErrorMessage)
	s.authService.AssertExpectations(s.T())
}

func TestUserHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}
