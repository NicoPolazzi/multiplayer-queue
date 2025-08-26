package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/gen/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type UserHandlerTestSuite struct {
	suite.Suite
	router      *gin.Engine
	mockGateway *httptest.Server
	handler     *UserHandler
}

func (s *UserHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.router = gin.Default()
	s.router.LoadHTMLGlob("../../web/templates/*")
}

func (s *UserHandlerTestSuite) AfterTest() {
	if s.mockGateway != nil {
		s.mockGateway.Close()
	}
}

func (s *UserHandlerTestSuite) setupMockGateway(handler http.HandlerFunc) {
	s.mockGateway = httptest.NewServer(handler)
	s.handler = NewUserHandler(s.mockGateway.URL)
}

func (s *UserHandlerTestSuite) TestPerformLoginSuccess() {
	s.setupMockGateway(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		mockResponse := &auth.LoginUserResponse{
			Token: "mock-jwt-token",
			User:  &auth.User{Id: 1, Username: "testuser"},
		}
		err := json.NewEncoder(w).Encode(mockResponse)
		if err != nil {
			s.Fail("Failed to encoding response")
		}
	})
	s.router.POST("/user/login", s.handler.PerformLogin)

	formData := url.Values{}
	formData.Set("username", "testuser")
	formData.Set("password", "password123")
	req, _ := http.NewRequest(http.MethodPost, "/user/login", strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusSeeOther, w.Code)
	s.Equal("/", w.Header().Get("Location"))
	s.Contains(w.Header().Get("Set-Cookie"), "token=mock-jwt-token")
}

func (s *UserHandlerTestSuite) TestPerformLoginFailureRendersLoginPageWithError() {
	s.setupMockGateway(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	s.router.POST("/user/login", s.handler.PerformLogin)

	formData := url.Values{}
	formData.Set("username", "testuser")
	formData.Set("password", "wrongpassword")
	req, _ := http.NewRequest(http.MethodPost, "/user/login", strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusUnauthorized, w.Code)

	htmlBody := w.Body.String()

	expectedErrorMessage := "Login Failed: Invalid username or password."
	s.Contains(htmlBody, expectedErrorMessage, "HTML body should contain the error message")
	s.Contains(htmlBody, "<form class=\"form\" action=\"/user/login\" method=\"POST\">", "HTML body should contain the login form")
	s.Contains(htmlBody, "<input type=\"text\" class=\"form-control\" id=\"username\" name=\"username\"", "HTML body should contain the username field")
}

func (s *UserHandlerTestSuite) TestPerformRegistrationSuccess() {
	s.setupMockGateway(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	s.router.POST("/user/register", s.handler.PerformRegistration)
	formData := url.Values{}
	formData.Set("username", "newuser")
	formData.Set("password", "password123")
	req, _ := http.NewRequest(http.MethodPost, "/user/register", strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusSeeOther, w.Code)
	s.Equal("/user/login", w.Header().Get("Location"))
}

func (s *UserHandlerTestSuite) TestPerformLogout() {
	s.setupMockGateway(nil)
	s.router.GET("/user/logout", s.handler.PerformLogout)

	req, _ := http.NewRequest(http.MethodGet, "/user/logout", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusSeeOther, w.Code)
	s.Equal("/", w.Header().Get("Location"))
	s.Contains(w.Header().Get("Set-Cookie"), "Max-Age=0")
}

func TestUserHandler(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}
