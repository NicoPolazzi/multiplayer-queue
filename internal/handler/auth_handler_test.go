package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	UserFixtureUsername string = "test"
	UserFixturePassword string = "password123"
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

type AuthHandlerTestSuite struct {
	suite.Suite
	mockService *MockAuthService
	handler     *AuthHandler
	recorder    *httptest.ResponseRecorder
	context     *gin.Context
}

func (s *AuthHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.recorder = httptest.NewRecorder()
	s.context, _ = gin.CreateTestContext(s.recorder)

	s.mockService = new(MockAuthService)
	s.handler = NewAuthHandler(s.mockService)
}

func (s *AuthHandlerTestSuite) TestRegisterHandler() {
	tests := []struct {
		name         string
		serviceError error
		expectedCode int
		expectedBody string
	}{
		{
			name:         "successful registration",
			serviceError: nil,
			expectedCode: http.StatusOK,
			expectedBody: `{"status": "success", "message": "User registered successfully"}`,
		},
		{
			name:         "username taken",
			serviceError: service.ErrUsernameTaken,
			expectedCode: http.StatusConflict,
			expectedBody: `{"status": "error", "message": "Username already taken"}`,
		},
		{
			name:         "server error",
			serviceError: errors.New("internal error"),
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"status": "error", "message": "Internal server error"}`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			s.setupRequest("POST", "/auth/register", gin.H{
				"username": UserFixtureUsername,
				"password": UserFixturePassword,
			})
			s.mockService.On("Register", UserFixtureUsername, UserFixturePassword).Return(tt.serviceError)

			s.handler.Register(s.context)

			assert.Equal(s.T(), tt.expectedCode, s.recorder.Code)
			assert.JSONEq(s.T(), tt.expectedBody, s.recorder.Body.String())
			s.mockService.AssertExpectations(s.T())
		})
	}
}

func (s *AuthHandlerTestSuite) setupRequest(method, path string, body map[string]any) {
	requestBody, _ := json.Marshal(body)
	s.context.Request = httptest.NewRequest(method, path, bytes.NewBuffer(requestBody))
	s.context.Request.Header.Set("Content-Type", "application/json")
}

func (s *AuthHandlerTestSuite) TestLoginHandler() {
	testToken := "test-token"

	tests := []struct {
		name         string
		serviceToken string
		serviceError error
		expectedCode int
		expectedBody string
	}{
		{
			name:         "successful login",
			serviceToken: testToken,
			serviceError: nil,
			expectedCode: http.StatusOK,
			expectedBody: `{"status": "success", "token": "` + testToken + `"}`,
		},
		{
			name:         "invalid credentials",
			serviceToken: "",
			serviceError: service.ErrInvalidCredentials,
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"status": "error", "message": "Invalid username or password"}`,
		},
		{
			name:         "server error",
			serviceToken: "",
			serviceError: errors.New("some internal error"),
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"status": "error", "message": "Internal server error"}`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			s.setupRequest("POST", "/auth/login", gin.H{
				"username": UserFixtureUsername,
				"password": UserFixturePassword,
			})
			s.mockService.On("Login", UserFixtureUsername, UserFixturePassword).Return(tt.serviceToken, tt.serviceError)

			s.handler.Login(s.context)

			assert.Equal(s.T(), tt.expectedCode, s.recorder.Code)
			assert.JSONEq(s.T(), tt.expectedBody, s.recorder.Body.String())
			s.mockService.AssertExpectations(s.T())
		})
	}
}

func TestAuthHandlerSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}
