package handler

import (
	"errors"
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
	request := bindAuthRequest(c)

	err := h.authService.Register(request.Username, request.Password)
	if err != nil {
		if errors.Is(err, service.ErrUsernameTaken) {
			c.JSON(http.StatusConflict, gin.H{"status": "error", "message": "Username already taken"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "User registered successfully"})
}

func bindAuthRequest(c *gin.Context) AuthRequest {
	var request AuthRequest
	c.ShouldBindJSON(&request)
	return request
}

func (h *AuthHandler) Login(c *gin.Context) {
	request := bindAuthRequest(c)

	token, err := h.authService.Login(request.Username, request.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Invalid username or password"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "token": token})
}
