package handlers

import (
	"errors"
	"net/http"

	"github.com/NicoPolazzi/multiplayer-queue/gen/auth"
	"github.com/NicoPolazzi/multiplayer-queue/internal/gateway"
	"github.com/gin-gonic/gin"
)

const (
	LoginPageFilename    = "login.html"
	RegisterPageFilename = "register.html"
)

type UserHandler struct {
	authClient *gateway.AuthGatewayClient
}

func NewUserHandler(client *gateway.AuthGatewayClient) *UserHandler {
	return &UserHandler{authClient: client}
}

func (h *UserHandler) ShowIndexPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"is_logged_in": c.GetBool("is_logged_in"),
		"lobbies":      c.MustGet("lobbies"),
		"username":     c.GetString("username"),
	})
}

func (h *UserHandler) ShowLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, LoginPageFilename, nil)
}

func (h *UserHandler) ShowRegisterPage(c *gin.Context) {
	c.HTML(http.StatusOK, RegisterPageFilename, nil)
}

func (h *UserHandler) PerformLogin(c *gin.Context) {
	loginReq := &auth.LoginUserRequest{
		Username: c.PostForm("username"),
		Password: c.PostForm("password"),
	}

	loginResponse, err := h.authClient.Login(c.Request.Context(), loginReq)
	if err != nil {
		var apiErr *gateway.APIError
		if errors.As(err, &apiErr) && (apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusNotFound) {
			c.HTML(apiErr.StatusCode, LoginPageFilename, gin.H{
				"ErrorTitle":   "Login Failed",
				"ErrorMessage": "Invalid username or password.",
			})
			return
		}

		c.HTML(http.StatusInternalServerError, LoginPageFilename, gin.H{
			"ErrorTitle":   "Service Error",
			"ErrorMessage": "The authentication service is currently unavailable.",
		})
		return
	}

	c.SetCookie("token", loginResponse.Token, 3600, "/", "", false, true)
	c.Redirect(http.StatusSeeOther, "/")
}

func (h *UserHandler) PerformRegistration(c *gin.Context) {
	regReq := &auth.RegisterUserRequest{
		Username: c.PostForm("username"),
		Password: c.PostForm("password"),
	}

	err := h.authClient.Register(c.Request.Context(), regReq)
	if err != nil {
		var apiErr *gateway.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusConflict {
			c.HTML(apiErr.StatusCode, RegisterPageFilename, gin.H{
				"ErrorTitle":   "Registration Failed",
				"ErrorMessage": "That username is already taken. Please choose another one.",
			})
			return
		}
		c.HTML(http.StatusInternalServerError, RegisterPageFilename, gin.H{
			"ErrorTitle":   "Registration Failed",
			"ErrorMessage": "An unexpected error occurred. Please try again.",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/user/login")
}

func (h *UserHandler) PerformLogout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.Redirect(http.StatusSeeOther, "/")
}
