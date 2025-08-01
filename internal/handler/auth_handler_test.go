package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestRegisterHandlerSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	requestBody, _ := json.Marshal(gin.H{"username": "testuser", "password": "password123"})

	ctx.Request = httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(requestBody))
	ctx.Request.Header.Set("Content-Type", "application/json")

	service := new(MockAuthService)
	service.On("Register", "testuser", "password123").Return(nil)
	handler := NewAuthHandler(service)

	handler.Register(ctx)

	assert.EqualValues(t, http.StatusOK, w.Code)
	service.AssertExpectations(t)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
}
