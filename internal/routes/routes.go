package routes

import (
	"github.com/NicoPolazzi/multiplayer-queue/internal/handlers"
	"github.com/NicoPolazzi/multiplayer-queue/internal/middleware"
	"github.com/NicoPolazzi/multiplayer-queue/internal/token"
	"github.com/gin-gonic/gin"
)

type RoutesManager struct {
	handler      *handlers.UserHandler
	tokenManager *token.TokenManager
}

func NewRoutes(handler *handlers.UserHandler, tokenManager *token.TokenManager) *RoutesManager {
	return &RoutesManager{handler: handler, tokenManager: tokenManager}
}

func (m *RoutesManager) InitializeRoutes(router *gin.Engine) {
	router.Use(middleware.CheckUser(*m.tokenManager))
	router.GET("/", handlers.ShowIndexPage)

	userRoutes := router.Group("/user")
	userRoutes.Use(middleware.EnsureNotLoggedIn())
	userRoutes.GET("/register", m.handler.ShowLRegisterPage)
	userRoutes.POST("/register", m.handler.PerformRegistration)
	userRoutes.GET("/login", m.handler.ShowLoginPage)
	userRoutes.POST("/login", m.handler.PerformLogin)

	protected := router.Group("/")
	protected.Use(middleware.EnsureLoggedIn())
	protected.GET("/user/logout", m.handler.PerformLogout)

}
