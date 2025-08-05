package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
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
	userFixtureUsername string = "test"
	userFixturePassword string = "password123"
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
	router      *gin.Engine
}

func (s *AuthHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.recorder = httptest.NewRecorder()
	s.context, s.router = gin.CreateTestContext(s.recorder)
	s.mockService = new(MockAuthService)
	s.handler = NewAuthHandler(s.mockService)
}

func (s *AuthHandlerTestSuite) TestRegisterHandler() {
	tests := []struct {
		name             string
		serviceError     error
		expectedCode     int
		expectedRedirect string
		expectedError    string
	}{
		{
			name:             "successful registration",
			serviceError:     nil,
			expectedCode:     http.StatusSeeOther,
			expectedRedirect: "/login",
		},
		{
			name:          "username taken",
			serviceError:  service.ErrUsernameTaken,
			expectedCode:  http.StatusConflict,
			expectedError: "Username already taken",
		},
		{
			name:          "server error",
			serviceError:  errors.New("internal error"),
			expectedCode:  http.StatusInternalServerError,
			expectedError: "something went wrong",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			s.setupRequest("POST", "/auth/register", gin.H{
				"username": userFixtureUsername,
				"password": userFixturePassword,
			})
			s.mockService.On("Register", userFixtureUsername, userFixturePassword).Return(tt.serviceError)

			s.handler.Register(s.context)

			if tt.expectedRedirect != "" {
				assert.Equal(s.T(), tt.expectedCode, s.recorder.Code)
				assert.Equal(s.T(), tt.expectedRedirect, s.recorder.Header().Get("Location"))
			} else {
				// Failure: an HTML response with an error message
				assert.Equal(s.T(), tt.expectedCode, s.recorder.Code)
				assert.Contains(s.T(), s.recorder.Body.String(), tt.expectedError)
			}
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
		name             string
		serviceToken     string
		serviceError     error
		expectedCode     int
		expectedRedirect string
		expectedError    string
	}{
		{
			name:             "successful login",
			serviceToken:     testToken,
			serviceError:     nil,
			expectedCode:     http.StatusSeeOther,
			expectedRedirect: "/dashboard",
		},
		{
			name:          "invalid credentials",
			serviceToken:  "",
			serviceError:  service.ErrInvalidCredentials,
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid username or password",
		},
		{
			name:          "server error",
			serviceToken:  "",
			serviceError:  errors.New("some internal error"),
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Something went wrong",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			s.setupRequest("POST", "/auth/login", gin.H{
				"username": userFixtureUsername,
				"password": userFixturePassword,
			})
			s.mockService.On("Login", userFixtureUsername, userFixturePassword).Return(tt.serviceToken, tt.serviceError)

			s.handler.Login(s.context)

			if tt.expectedRedirect != "" {
				assert.Equal(s.T(), tt.expectedCode, s.recorder.Code)
				assert.Equal(s.T(), tt.expectedRedirect, s.recorder.Header().Get("Location"))
				// Check the cookie
				cookies := s.recorder.Result().Cookies()
				found := false
				for _, cookie := range cookies {
					if cookie.Name == "jwt" && cookie.Value == testToken {
						found = true
						break
					}
				}
				assert.True(s.T(), found, "jwt cookie was not set")
			} else {
				assert.Equal(s.T(), tt.expectedCode, s.recorder.Code)
				assert.Contains(s.T(), s.recorder.Body.String(), tt.expectedError)
			}
			s.mockService.AssertExpectations(s.T())
		})
	}
}

func (s *AuthHandlerTestSuite) TestShowLogin() {
	templ := template.Must(template.New("login.html").Parse(`
        <!DOCTYPE html>
        <html>
        <head>
            <title>Login</title>
        </head>
        <body>
            <h2>Login</h2>
            <form action="/login" method="POST">
                <input name="username" placeholder="Username" required><br>
                <input type="password" name="password" placeholder="Password" required><br>
                <button type="submit">Login</button>
            </form>
        </body>
        </html>
    `))
	s.router.SetHTMLTemplate(templ)
	s.handler.ShowLogin(s.context)
	assert.Equal(s.T(), http.StatusOK, s.recorder.Code)
	assert.Contains(s.T(), s.recorder.Body.String(), "Login")
}

func (s *AuthHandlerTestSuite) TestShowRegister() {
	templ := template.Must(template.New("register.html").Parse(`
        <!DOCTYPE html>
        <html>
        <head>
            <title>Register</title>
        </head>
        <body>
            <h2>Register</h2>
            <form action="/register" method="POST">
                <input name="username" placeholder="Username" required><br>
                <input type="password" name="password" placeholder="Password" required><br>
                <button type="submit">Register</button>
            </form>
        </body>
        </html>
    `))
	s.router.SetHTMLTemplate(templ)
	s.handler.ShowRegister(s.context)
	assert.Equal(s.T(), http.StatusOK, s.recorder.Code)
	assert.Contains(s.T(), s.recorder.Body.String(), "Register")
}

func TestAuthHandlerSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}
