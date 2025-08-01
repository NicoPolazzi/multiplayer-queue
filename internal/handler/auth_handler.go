package handler

import (
	"net/http"

	"github.com/NicoPolazzi/multiplayer-queue/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

type AuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var request AuthRequest
	c.ShouldBindJSON(&request)

	err := h.authService.Register(request.Username, request.Password)
	if err == service.ErrUsernameTaken {
		c.JSON(http.StatusConflict, gin.H{"status": "error", "message": "Username already taken"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "User registered successfully"})
}
