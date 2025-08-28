package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/gen/auth"
	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/NicoPolazzi/multiplayer-queue/internal/gateway"
	"github.com/NicoPolazzi/multiplayer-queue/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/encoding/protojson"
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

type UserHandlerTestSuite struct {
	suite.Suite
	router           *gin.Engine
	mockAuthGateway  *httptest.Server
	mockLobbyGateway *httptest.Server
	handler          *UserHandler
	authClient       *gateway.AuthGatewayClient
	lobbyClient      *gateway.LobbyGatewayClient
	mockTokenManager *MockTokenManager
}

func (s *UserHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.router = gin.Default()
	s.router.LoadHTMLGlob("../../web/templates/*")

	s.mockTokenManager = new(MockTokenManager)

	authMiddleware := middleware.NewAuthMiddleware(s.mockTokenManager)
	s.router.Use(authMiddleware.CheckUser())
}

func (s *UserHandlerTestSuite) TearDownTest() {
	if s.mockAuthGateway != nil {
		s.mockAuthGateway.Close()
	}
	if s.mockLobbyGateway != nil {
		s.mockLobbyGateway.Close()
	}
}

func (s *UserHandlerTestSuite) setup(authHandler, lobbyHandler http.HandlerFunc) {
	if authHandler != nil {
		s.mockAuthGateway = httptest.NewServer(authHandler)
		s.authClient = gateway.NewAuthGatewayClient(s.mockAuthGateway.URL)
	}
	if lobbyHandler != nil {
		s.mockLobbyGateway = httptest.NewServer(lobbyHandler)
		s.lobbyClient = gateway.NewLobbyGatewayClient(s.mockLobbyGateway.URL)
	}
	s.handler = NewUserHandler(s.authClient, s.lobbyClient)
}

func (s *UserHandlerTestSuite) TestShowIndexPageAsLoggedInUser() {
	s.setup(nil, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := &lobby.ListAvailableLobbiesResponse{Lobbies: []*lobby.Lobby{{Name: "Fun Lobby"}}}
		body, _ := protojson.Marshal(resp)
		w.Write(body)
	})
	s.router.GET("/", s.handler.ShowIndexPage)
	s.mockTokenManager.On("Validate", "valid-token").Return("testuser", nil)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "valid-token"})

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "Welcome back, testuser!")
	s.Contains(w.Body.String(), "Fun Lobby")
	s.mockTokenManager.AssertExpectations(s.T())
}

func (s *UserHandlerTestSuite) TestShowIndexPageLobbyServiceFails() {
	s.setup(nil, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	s.router.GET("/", s.handler.ShowIndexPage)
	s.mockTokenManager.On("Validate", "valid-token").Return("testuser", nil)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "valid-token"})

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "Lobby Service Error")
	s.Contains(w.Body.String(), "Could not retrieve the list of available lobbies.")
	s.mockTokenManager.AssertExpectations(s.T())
}

func (s *UserHandlerTestSuite) TestShowLoginPage() {
	s.setup(nil, nil) // No gateway calls are needed for this handler
	s.router.GET(LoginPath, s.handler.ShowLoginPage)

	req, _ := http.NewRequest(http.MethodGet, LoginPath, nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "<title>Login</title>", "Expected the login page title")
	s.Contains(w.Body.String(), "action=\"/user/login\"", "Expected the login form")
}

func (s *UserHandlerTestSuite) TestShowRegisterPage() {
	s.setup(nil, nil) // No gateway calls are needed for this handler
	s.router.GET("/user/register", s.handler.ShowRegisterPage)

	req, _ := http.NewRequest(http.MethodGet, "/user/register", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "<title>Register</title>", "Expected the register page title")
	s.Contains(w.Body.String(), "action=\"/user/register\"", "Expected the register form")
}

func (s *UserHandlerTestSuite) TestPerformLoginSuccess() {
	s.setup(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := &auth.LoginUserResponse{Token: "mock-jwt-token"}
		body, _ := protojson.Marshal(resp)
		w.Write(body)
	}, nil)
	s.router.POST(LoginPath, s.handler.PerformLogin)

	formData := url.Values{"username": {"testuser"}, "password": {"password123"}}
	req, _ := http.NewRequest(http.MethodPost, LoginPath, strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusSeeOther, w.Code)
	s.Equal("/", w.Header().Get("Location"))
	s.Contains(w.Header().Get("Set-Cookie"), "token=mock-jwt-token")
}

func (s *UserHandlerTestSuite) TestPerformLoginGatewayGenericFailure() {
	s.setup(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}, nil)
	s.router.POST("/user/login", s.handler.PerformLogin)

	formData := url.Values{"username": {"testuser"}, "password": {"password123"}}
	req, _ := http.NewRequest(http.MethodPost, "/user/login", strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), "Service Error")
	s.Contains(w.Body.String(), "The authentication service is currently unavailable.")
}

func (s *UserHandlerTestSuite) TestPerformLoginFailsWithInvalidCredentials() {
	// Setup the mock gateway to return a 401 Unauthorized status.
	// This will cause the client to return an APIError with StatusCode 401.
	s.setup(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}, nil)
	s.router.POST("/user/login", s.handler.PerformLogin)

	formData := url.Values{"username": {"testuser"}, "password": {"wrongpassword"}}
	req, _ := http.NewRequest(http.MethodPost, "/user/login", strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusUnauthorized, w.Code)
	s.Contains(w.Body.String(), "Login Failed")
	s.Contains(w.Body.String(), "Invalid username or password.")
}

func (s *UserHandlerTestSuite) TestPerformRegistrationSuccess() {
	s.setup(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, nil)
	s.router.POST("/user/register", s.handler.PerformRegistration)

	formData := url.Values{"username": {"newuser"}, "password": {"password123"}}
	req, _ := http.NewRequest(http.MethodPost, "/user/register", strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusSeeOther, w.Code)
	s.Equal(LoginPath, w.Header().Get("Location"))
}

func (s *UserHandlerTestSuite) TestPerformRegistrationFailsOnUsernameConflict() {
	s.setup(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}, nil)
	s.router.POST("/user/register", s.handler.PerformRegistration)

	formData := url.Values{"username": {"existinguser"}, "password": {"password123"}}
	req, _ := http.NewRequest(http.MethodPost, "/user/register", strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusConflict, w.Code)
	s.Contains(w.Body.String(), "Registration Failed")
	s.Contains(w.Body.String(), "That username is already taken.")
}

func (s *UserHandlerTestSuite) TestPerformRegistrationFailsOnGatewayError() {
	s.setup(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}, nil)
	s.router.POST("/user/register", s.handler.PerformRegistration)

	formData := url.Values{"username": {"anyuser"}, "password": {"password123"}}
	req, _ := http.NewRequest(http.MethodPost, "/user/register", strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), "Registration Failed")
	s.Contains(w.Body.String(), "An unexpected error occurred.")
}

func (s *UserHandlerTestSuite) TestPerformLogout() {
	s.setup(nil, nil) // No gateway calls needed
	s.router.GET("/user/logout", s.handler.PerformLogout)

	req, _ := http.NewRequest(http.MethodGet, "/user/logout", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusSeeOther, w.Code)
	s.Equal("/", w.Header().Get("Location"))
	s.Contains(w.Header().Get("Set-Cookie"), "token=;")
}

func TestUserHandler(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}
