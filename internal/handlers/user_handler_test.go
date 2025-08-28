package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/gen/auth"
	"github.com/NicoPolazzi/multiplayer-queue/internal/gateway"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/encoding/protojson"
)

type UserHandlerTestSuite struct {
	suite.Suite
	router      *gin.Engine
	mockGateway *httptest.Server
	handler     *UserHandler
	authClient  *gateway.AuthGatewayClient
}

func (s *UserHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.router = gin.Default()
	// Adjust the path according to your project structure
	s.router.LoadHTMLGlob("../web/templates/*")
}

func (s *UserHandlerTestSuite) AfterTest() {
	if s.mockGateway != nil {
		s.mockGateway.Close()
	}
}

func (s *UserHandlerTestSuite) setupMockGateway(handler http.HandlerFunc) {
	s.mockGateway = httptest.NewServer(handler)
	s.authClient = gateway.NewAuthGatewayClient(s.mockGateway.URL)
	s.handler = NewUserHandler(s.authClient, nil)
}

func (s *UserHandlerTestSuite) TestPerformLoginSuccess() {
	s.setupMockGateway(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		mockResponse := &auth.LoginUserResponse{
			Token: "mock-jwt-token",
			User:  &auth.User{Id: 1, Username: "testuser"},
		}
		body, err := protojson.Marshal(mockResponse)
		s.Require().NoError(err)
		w.Write(body)
	})
	s.router.POST("/user/login", s.handler.PerformLogin)

	formData := url.Values{"username": {"testuser"}, "password": {"password123"}}
	req, _ := http.NewRequest(http.MethodPost, "/user/login", strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusSeeOther, w.Code)
	s.Equal("/", w.Header().Get("Location"))
	s.Contains(w.Header().Get("Set-Cookie"), "token=mock-jwt-token")
}

func (s *UserHandlerTestSuite) TestPerformLoginFailure() {
	s.setupMockGateway(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	s.router.POST("/user/login", s.handler.PerformLogin)

	formData := url.Values{"username": {"testuser"}, "password": {"wrongpassword"}}
	req, _ := http.NewRequest(http.MethodPost, "/user/login", strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusUnauthorized, w.Code)
	s.Contains(w.Body.String(), "Invalid username or password.")
}

func (s *UserHandlerTestSuite) TestPerformRegistrationSuccess() {
	// 1. Setup the mock gateway to return a 200 OK for a successful registration.
	s.setupMockGateway(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	s.router.POST("/user/register", s.handler.PerformRegistration)

	// 2. Create the form data for the new user.
	formData := url.Values{"username": {"newuser"}, "password": {"password123"}}
	req, _ := http.NewRequest(http.MethodPost, "/user/register", strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// 3. Assert that the user is redirected to the login page on success.
	s.Equal(http.StatusSeeOther, w.Code)
	s.Equal("/user/login", w.Header().Get("Location"))
}

func (s *UserHandlerTestSuite) TestPerformRegistrationFailureConflict() {
	// 1. Setup the mock gateway to return a 409 Conflict for a duplicate username.
	s.setupMockGateway(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	})
	s.router.POST("/user/register", s.handler.PerformRegistration)

	// 2. Create the form data.
	formData := url.Values{"username": {"existinguser"}, "password": {"password123"}}
	req, _ := http.NewRequest(http.MethodPost, "/user/register", strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// 3. Assert that the register page is re-rendered with the correct error message.
	s.Equal(http.StatusConflict, w.Code)
	s.Contains(w.Body.String(), "That username is already taken.")
}

func (s *UserHandlerTestSuite) TestPerformLogout() {
	// Logout is simple and doesn't need to contact the gateway.
	s.setupMockGateway(nil)
	s.router.GET("/user/logout", s.handler.PerformLogout)

	req, _ := http.NewRequest(http.MethodGet, "/user/logout", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Assert that the user is redirected and the cookie is expired.
	s.Equal(http.StatusSeeOther, w.Code)
	s.Equal("/", w.Header().Get("Location"))
	s.Contains(w.Header().Get("Set-Cookie"), "token=;")
	s.Contains(w.Header().Get("Set-Cookie"), "Max-Age=0") // A more precise check for cookie deletion.
}

func TestUserHandler(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}
