package handlers

import (
	"errors"
	"net/http"

	"github.com/NicoPolazzi/multiplayer-queue/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	indexPageFilename        string = "index.html"
	registerPageFilename     string = "register.html"
	loginPageFilename        string = "login.html"
	RegistrationErrorMessage string = "Registration Failed"
	LoginErrorMessage        string = "Login Failed"
	IndexPagePath            string = "/"
	LoginPagePath            string = "/user/login"
)

type AuthRequest struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type UserHandler struct {
	authService service.AuthService
}

func NewUserHandler(authService service.AuthService) *UserHandler {
	return &UserHandler{authService: authService}
}

func (h *UserHandler) ShowIndexPage(c *gin.Context) {
	isLoggedIn, _ := c.Get("is_logged_in")
	username, _ := c.Get("username")
	lobbies, _ := c.Get("lobbies")

	c.HTML(http.StatusOK, indexPageFilename, gin.H{
		"title":        "Home Page",
		"is_logged_in": isLoggedIn,
		"username":     username,
		"lobbies":      lobbies,
	})
}

func (h *UserHandler) ShowLRegisterPage(c *gin.Context) {
	c.HTML(http.StatusOK, registerPageFilename, gin.H{"title": "Register"})
}

func (h *UserHandler) PerformRegistration(c *gin.Context) {
	var request AuthRequest
	if err := c.ShouldBind(&request); err != nil {
		c.HTML(http.StatusBadRequest, registerPageFilename, gin.H{
			"ErrorTitle":   RegistrationErrorMessage,
			"ErrorMessage": err.Error()})
		return
	}

	err := h.authService.Register(request.Username, request.Password)
	if err != nil {
		if errors.Is(err, service.ErrUsernameTaken) {
			c.HTML(http.StatusBadRequest, registerPageFilename, gin.H{
				"ErrorTitle":   RegistrationErrorMessage,
				"ErrorMessage": err.Error()})
		} else {
			c.HTML(http.StatusInternalServerError, registerPageFilename, gin.H{
				"ErrorTitle":   RegistrationErrorMessage,
				"ErrorMessage": err.Error()})
		}
		return
	}

	c.Redirect(http.StatusSeeOther, "/")
}

func (h *UserHandler) ShowLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, loginPageFilename, gin.H{
		"title": "Login",
	})
}

func (h *UserHandler) PerformLogin(c *gin.Context) {
	var request AuthRequest
	if err := c.ShouldBind(&request); err != nil {
		c.HTML(http.StatusBadRequest, loginPageFilename, gin.H{
			"ErrorTitle":   LoginErrorMessage,
			"ErrorMessage": "Missing credentials"})
		return
	}

	token, err := h.authService.Login(request.Username, request.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.HTML(http.StatusUnauthorized, loginPageFilename, gin.H{
				"ErrorTitle":   LoginErrorMessage,
				"ErrorMessage": err.Error()})
		} else {
			c.HTML(http.StatusInternalServerError, loginPageFilename, gin.H{
				"ErrorTitle":   LoginErrorMessage,
				"ErrorMessage": err.Error()})
			return
		}
		return
	}

	c.SetCookie("jwt", token, 3600*24, "/", "", false, true)
	c.Set("is_logged_in", true)
	c.Redirect(http.StatusSeeOther, "/")
}

func (h *UserHandler) PerformLogout(c *gin.Context) {
	c.SetCookie("jwt", "", -1, "/", "", false, true)
	c.Set("is_logged_in", false)
	c.Redirect(http.StatusSeeOther, "/")
}
