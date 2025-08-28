package routes

import (
	"github.com/NicoPolazzi/multiplayer-queue/internal/handlers"
	"github.com/NicoPolazzi/multiplayer-queue/internal/middleware"
	"github.com/gin-gonic/gin"
)

type RoutesManager struct {
	userHandler    *handlers.UserHandler
	lobbyHandler   *handlers.LobbyHandler
	authMiddleware *middleware.AuthMiddleware
}

func NewRoutes(userHandler *handlers.UserHandler,
	lobbyHandler *handlers.LobbyHandler,
	authMiddleware *middleware.AuthMiddleware) *RoutesManager {
	return &RoutesManager{userHandler: userHandler, lobbyHandler: lobbyHandler, authMiddleware: authMiddleware}
}

func (m *RoutesManager) InitializeRoutes(router *gin.Engine) {
	router.Use(m.authMiddleware.CheckUser())

	// Routes for guest users
	userRoutes := router.Group("/user")
	userRoutes.Use(middleware.EnsureNotLoggedIn())
	{
		userRoutes.GET("/register", m.userHandler.ShowRegisterPage)
		userRoutes.POST("/register", m.userHandler.PerformRegistration)
		userRoutes.GET("/login", m.userHandler.ShowLoginPage)
		userRoutes.POST("/login", m.userHandler.PerformLogin)
	}

	// Routes for logged users
	protected := router.Group("/")
	protected.Use(middleware.EnsureLoggedIn())
	{
		protected.POST("/lobbies/create", m.lobbyHandler.CreateLobby)
		protected.POST("/lobbies/:lobby_id/join", m.lobbyHandler.JoinLobby)
		protected.GET("/lobbies/:lobby_id", m.lobbyHandler.GetLobbyPage)

		protected.PUT("/api/v1/lobbies/:lobby_id/finish", m.lobbyHandler.FinishLobby)

		protected.GET("/user/logout", m.userHandler.PerformLogout)
	}

	router.GET("/", m.userHandler.ShowIndexPage)
}
