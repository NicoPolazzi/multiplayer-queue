package routes

import (
	"github.com/NicoPolazzi/multiplayer-queue/internal/handlers"
	"github.com/NicoPolazzi/multiplayer-queue/internal/middleware"
	"github.com/gin-gonic/gin"
)

type RoutesManager struct {
	handler        *handlers.UserHandler
	authMiddleware *middleware.AuthMiddleware
}

func NewRoutes(handler *handlers.UserHandler, authMiddleware *middleware.AuthMiddleware) *RoutesManager {
	return &RoutesManager{handler: handler, authMiddleware: authMiddleware}
}

func (m *RoutesManager) InitializeRoutes(router *gin.Engine) {
	router.Use(m.authMiddleware.CheckUser())
	router.GET("/", middleware.LoadLobbies(), m.handler.ShowIndexPage)

	userRoutes := router.Group("/user")
	userRoutes.Use(middleware.EnsureNotLoggedIn())
	userRoutes.GET("/register", m.handler.ShowLRegisterPage)
	userRoutes.POST("/register", m.handler.PerformRegistration)
	userRoutes.GET("/login", m.handler.ShowLoginPage)
	userRoutes.POST("/login", m.handler.PerformLogin)

	protected := router.Group("/")
	protected.Use(middleware.EnsureLoggedIn())
	protected.GET("/user/logout", m.handler.PerformLogout)
	protected.POST("/lobbies/create", handlers.CreateLobby)
}
