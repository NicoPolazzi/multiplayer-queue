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

func TestRegisterHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

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
			serviceError: errors.New("some internal error"),
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"status": "error", "message": "Internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			requestBody, _ := json.Marshal(gin.H{
				"username": UserFixtureUsername,
				"password": UserFixturePassword,
			})

			ctx.Request = httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(requestBody))
			ctx.Request.Header.Set("Content-Type", "application/json")

			mockService := new(MockAuthService)
			mockService.On("Register", UserFixtureUsername, UserFixturePassword).Return(tt.serviceError)
			handler := NewAuthHandler(mockService)

			handler.Register(ctx)

			var response map[string]any
			err := json.Unmarshal([]byte(w.Body.String()), &response)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}
