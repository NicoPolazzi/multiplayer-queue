package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/NicoPolazzi/multiplayer-queue/gen/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	LoginPageFilename    = "login.html"
	RegisterPageFilename = "register.html"
)

type UserHandler struct {
	gatewayBaseURL string
}

func NewUserHandler(gatewayBaseURL string) *UserHandler {
	return &UserHandler{gatewayBaseURL: gatewayBaseURL}
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
	username := c.PostForm("username")
	password := c.PostForm("password")

	loginReq := &auth.LoginUserRequest{
		Username: username,
		Password: password,
	}
	reqBody, _ := protojson.Marshal(loginReq)

	resp, err := http.Post(h.gatewayBaseURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		c.HTML(http.StatusInternalServerError, LoginPageFilename, gin.H{
			"ErrorTitle":   "Service Error",
			"ErrorMessage": "Could not contact the authentication service.",
		})
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.HTML(resp.StatusCode, LoginPageFilename, gin.H{
			"ErrorTitle":   "Login Failed",
			"ErrorMessage": "Invalid username or password.",
		})
		return
	}

	var loginResponse auth.LoginUserResponse
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &loginResponse); err != nil {
		c.HTML(http.StatusInternalServerError, LoginPageFilename, gin.H{
			"ErrorTitle":   "Server Error",
			"ErrorMessage": "Failed to process the login response.",
		})
		return
	}

	c.SetCookie("token", loginResponse.Token, 3600, "/", "", false, true)
	c.Redirect(http.StatusSeeOther, "/")
}

func (h *UserHandler) PerformRegistration(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	regReq := &auth.RegisterUserRequest{
		Username: username,
		Password: password,
	}
	reqBody, _ := protojson.Marshal(regReq)

	resp, err := http.Post(h.gatewayBaseURL+"/api/v1/auth/register", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		c.HTML(http.StatusInternalServerError, RegisterPageFilename, gin.H{
			"ErrorTitle":   "Service Error",
			"ErrorMessage": "Could not contact the auth service.",
		})
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	// On successful registration, gateway returns 200 OK.
	if resp.StatusCode == http.StatusOK {
		c.Redirect(http.StatusSeeOther, "/user/login")
		return
	}

	if resp.StatusCode == http.StatusConflict {
		c.HTML(resp.StatusCode, RegisterPageFilename, gin.H{
			"ErrorTitle":   "Registration Failed",
			"ErrorMessage": "That username is already taken. Please choose another one.",
		})
		return
	}

	// Handle all other errors generically.
	c.HTML(resp.StatusCode, RegisterPageFilename, gin.H{
		"ErrorTitle":   "Registration Failed",
		"ErrorMessage": "An unexpected error occurred. Please try again.",
	})
}

func (h *UserHandler) PerformLogout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.Redirect(http.StatusSeeOther, "/")
}
