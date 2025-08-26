package routes

import (
	"net/http"
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/internal/handlers"
	"github.com/NicoPolazzi/multiplayer-queue/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestInitializeRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	manager := NewRoutes(
		&handlers.UserHandler{},
		&handlers.LobbyHandler{},
		&middleware.AuthMiddleware{},
		&middleware.LobbyMiddleware{},
	)
	manager.InitializeRoutes(router)

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

	registeredRoutes := router.Routes()
	assert.Len(t, registeredRoutes, len(expectedRoutes), "The number of registered routes should match the expected count.")

	routeMap := make(map[string]bool)
	for _, route := range registeredRoutes {
		routeMap[route.Method+route.Path] = true
	}

	for _, expected := range expectedRoutes {
		key := expected.Method + expected.Path
		assert.True(t, routeMap[key], "Route %s %s was not registered.", expected.Method, expected.Path)
	}
}
