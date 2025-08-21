package routes

import (
	"github.com/NicoPolazzi/multiplayer-queue/internal/handlers"
	"github.com/NicoPolazzi/multiplayer-queue/internal/middleware"
	"github.com/gin-gonic/gin"
)

type RoutesManager struct {
	userHandler     *handlers.UserHandler
	lobbyHandler    *handlers.LobbyHandler
	authMiddleware  *middleware.AuthMiddleware
	lobbyMiddleware *middleware.LobbyMiddleware
}

func NewRoutes(userHandler *handlers.UserHandler,
	lobbyHandler *handlers.LobbyHandler,
	authMiddleware *middleware.AuthMiddleware,
	lobbyMiddleware *middleware.LobbyMiddleware) *RoutesManager {
	return &RoutesManager{userHandler: userHandler, lobbyHandler: lobbyHandler, authMiddleware: authMiddleware,
		lobbyMiddleware: lobbyMiddleware}
}

func (m *RoutesManager) InitializeRoutes(router *gin.Engine) {
	router.Use(m.authMiddleware.CheckUser())
	router.GET("/", m.lobbyMiddleware.LoadLobbies(), m.userHandler.ShowIndexPage)

	userRoutes := router.Group("/user")
	userRoutes.Use(middleware.EnsureNotLoggedIn())
	userRoutes.GET("/register", m.userHandler.ShowLRegisterPage)
	userRoutes.POST("/register", m.userHandler.PerformRegistration)
	userRoutes.GET("/login", m.userHandler.ShowLoginPage)
	userRoutes.POST("/login", m.userHandler.PerformLogin)

	protected := router.Group("/")
	protected.Use(middleware.EnsureLoggedIn())
	protected.GET("/user/logout", m.userHandler.PerformLogout)
	protected.POST("/lobbies/create", m.lobbyHandler.CreateLobby)
	protected.POST("/lobbies/:lobby_id/join", m.lobbyHandler.JoinLobby)
	protected.GET("/lobbies/:lobby_id", m.lobbyHandler.GetLobbyPage)
	protected.PUT("/api/v1/lobbies/:lobby_id/finish", m.lobbyHandler.FinishLobby)
}
