package handler

import (
	"errors"
	"net/http"

	"github.com/NicoPolazzi/multiplayer-queue/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	basePageFilename string = "base.html"
)

type AuthHandler struct {
	authService service.AuthService
}

type AuthRequest struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) ShowLogin(c *gin.Context) {
	c.HTML(http.StatusOK, basePageFilename, gin.H{
		"title":    "Login",
		"template": "login",
	})
}

func (h *AuthHandler) ShowRegister(c *gin.Context) {
	c.HTML(http.StatusOK, basePageFilename, gin.H{
		"title":    "Register",
		"template": "register",
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var request AuthRequest
	if err := c.ShouldBind(&request); err != nil {
		c.HTML(http.StatusBadRequest, basePageFilename, gin.H{"template": "register", "error": "Missing credentials"})
		return
	}

	err := h.authService.Register(request.Username, request.Password)
	if err != nil {
		if errors.Is(err, service.ErrUsernameTaken) {
			c.HTML(http.StatusConflict, basePageFilename, gin.H{"error": err.Error()})
		} else {
			c.HTML(http.StatusInternalServerError, basePageFilename, gin.H{"template": "register", "error": "something went wrong"})
		}
		return
	}

	c.Redirect(http.StatusSeeOther, "/login")
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request AuthRequest
	if err := c.ShouldBind(&request); err != nil {
		c.HTML(http.StatusBadRequest, basePageFilename, gin.H{"template": "login", "error": "Missing credentials"})
		return
	}

	token, err := h.authService.Login(request.Username, request.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.HTML(http.StatusUnauthorized, basePageFilename, gin.H{"template": "login", "error": err.Error()})
		} else {
			c.HTML(http.StatusInternalServerError, basePageFilename, gin.H{"template": "login", "error": "Something went wrong"})
		}
		return
	}

	c.SetCookie("jwt", token, 3600*24, "/", "", false, true)
	// Using 303 See Other instructs the browser to perform a GET request to /dashboard
	c.Redirect(http.StatusSeeOther, "/dashboard")
}
