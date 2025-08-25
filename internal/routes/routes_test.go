package routes

import (
	"net/http"
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/internal/handlers"
	"github.com/NicoPolazzi/multiplayer-queue/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

// We are not testing the logic of the handlers
// or middleware here, only that they are correctly assigned.
type mockUserHandler struct{}
type mockLobbyHandler struct{}
type mockAuthMiddleware struct{}
type mockLobbyMiddleware struct{}

func (h *mockUserHandler) ShowRegisterPage(c *gin.Context)    {}
func (h *mockUserHandler) PerformRegistration(c *gin.Context) {}
func (h *mockUserHandler) ShowLoginPage(c *gin.Context)       {}
func (h *mockUserHandler) PerformLogin(c *gin.Context)        {}
func (h *mockUserHandler) PerformLogout(c *gin.Context)       {}
func (h *mockUserHandler) ShowIndexPage(c *gin.Context)       {}

func (h *mockLobbyHandler) CreateLobby(c *gin.Context)  {}
func (h *mockLobbyHandler) JoinLobby(c *gin.Context)    {}
func (h *mockLobbyHandler) GetLobbyPage(c *gin.Context) {}
func (h *mockLobbyHandler) FinishLobby(c *gin.Context)  {}

func (m *mockAuthMiddleware) CheckUser() gin.HandlerFunc {
	return func(c *gin.Context) {}
}

func (m *mockLobbyMiddleware) LoadLobbies() gin.HandlerFunc {
	return func(c *gin.Context) {}
}

type RoutesManagerTestSuite struct {
	suite.Suite
	router              *gin.Engine
	mockUserHandler     *handlers.UserHandler
	mockLobbyHandler    *handlers.LobbyHandler
	mockAuthMiddleware  *middleware.AuthMiddleware
	mockLobbyMiddleware *middleware.LobbyMiddleware
}

func (s *RoutesManagerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)

	s.mockUserHandler = &handlers.UserHandler{}
	s.mockLobbyHandler = &handlers.LobbyHandler{}
	s.mockAuthMiddleware = &middleware.AuthMiddleware{}
	s.mockLobbyMiddleware = &middleware.LobbyMiddleware{}

	s.router = gin.New()
}

func (s *RoutesManagerTestSuite) TestInitializeRoutes() {
	manager := NewRoutes(
		s.mockUserHandler,
		s.mockLobbyHandler,
		s.mockAuthMiddleware,
		s.mockLobbyMiddleware,
	)
	manager.InitializeRoutes(s.router)

	registeredRoutes := s.router.Routes()

	expectedRoutes := []struct {
		Method string
		Path   string
	}{
		{http.MethodGet, "/user/register"},
		{http.MethodPost, "/user/register"},
		{http.MethodGet, "/user/login"},
		{http.MethodPost, "/user/login"},
		{http.MethodPost, "/lobbies/create"},
		{http.MethodPost, "/lobbies/:lobby_id/join"},
		{http.MethodGet, "/lobbies/:lobby_id"},
		{http.MethodPut, "/api/v1/lobbies/:lobby_id/finish"},
		{http.MethodGet, "/user/logout"},
		{http.MethodGet, "/"},
	}

	s.Len(registeredRoutes, len(expectedRoutes), "The number of registered routes should match the expected count.")

	for _, expected := range expectedRoutes {
		s.assertRouteExists(registeredRoutes, expected.Method, expected.Path)
	}
}

func (s *RoutesManagerTestSuite) assertRouteExists(routes gin.RoutesInfo, method, path string) {
	found := false
	for _, route := range routes {
		if route.Method == method && route.Path == path {
			found = true
			break
		}
	}
	s.True(found, "Route %s %s was not registered.", method, path)
}

func TestRoutes(t *testing.T) {
	suite.Run(t, new(RoutesManagerTestSuite))
}
