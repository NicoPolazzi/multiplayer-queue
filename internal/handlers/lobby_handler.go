package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/NicoPolazzi/multiplayer-queue/internal/gateway"
	"github.com/NicoPolazzi/multiplayer-queue/internal/middleware"
	"github.com/gin-gonic/gin"
)

const (
	indexPageFilename = "index.html"
	lobbyPageFilename = "lobby.html"
)

type LobbyHandler struct {
	lobbyClient *gateway.LobbyGatewayClient
}

func NewLobbyHandler(client *gateway.LobbyGatewayClient) *LobbyHandler {
	return &LobbyHandler{lobbyClient: client}
}

func (h *LobbyHandler) CreateLobby(c *gin.Context) {
	// Is ok to not check the existence of the user because this handler is protected with the authMiddleware
	user, _ := middleware.UserFromContext(c)

	lobbyName := c.PostForm("name")
	if lobbyName == "" {
		c.HTML(http.StatusBadRequest, indexPageFilename, gin.H{
			"ErrorTitle":   "Lobby Creation Failed",
			"ErrorMessage": "Lobby name cannot be empty.",
			"is_logged_in": true,
			"username":     user.Username,
		})
		return
	}

	createReq := &lobby.CreateLobbyRequest{
		Name:     lobbyName,
		Username: user.Username,
	}

	newLobby, err := h.lobbyClient.CreateLobby(c.Request.Context(), createReq)
	if err != nil {
		c.HTML(http.StatusInternalServerError, indexPageFilename, gin.H{
			"ErrorTitle":   "Lobby Creation Failed",
			"ErrorMessage": fmt.Sprintf("An unexpected error occurred while creating the lobby: %v.", err),
			"is_logged_in": true,
			"username":     user.Username,
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/lobbies/"+newLobby.LobbyId)
}

func (h *LobbyHandler) JoinLobby(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	lobbyID := c.Param("lobby_id")

	joinReq := &lobby.JoinLobbyRequest{
		LobbyId:  lobbyID,
		Username: user.Username,
	}

	err := h.lobbyClient.JoinLobby(c.Request.Context(), joinReq)
	if err != nil {
		c.HTML(http.StatusInternalServerError, indexPageFilename, gin.H{
			"ErrorTitle":   "Join Lobby Failed",
			"ErrorMessage": "An unexpected error occurred while joining the lobby.",
			"is_logged_in": true,
			"username":     user.Username,
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/lobbies/"+lobbyID)
}

func (h *LobbyHandler) GetLobbyPage(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	lobbyID := c.Param("lobby_id")

	lobbyData, err := h.lobbyClient.GetLobby(c.Request.Context(), lobbyID)
	if err != nil {
		var apiErr *gateway.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			c.HTML(http.StatusNotFound, lobbyPageFilename, gin.H{
				"ErrorTitle":   "Lobby Not Found",
				"ErrorMessage": "The lobby you are looking for does not exist.",
			})
			return
		}

		c.HTML(http.StatusInternalServerError, indexPageFilename, gin.H{
			"ErrorTitle":   "Error Fetching Lobby",
			"ErrorMessage": "The server is currently unavailable.",
		})
		return
	}

	c.HTML(http.StatusOK, lobbyPageFilename, gin.H{
		"lobby":        lobbyData,
		"is_logged_in": c.GetBool("is_logged_in"),
		"username":     user.Username,
	})
}

func (h *LobbyHandler) FinishLobby(c *gin.Context) {
	lobbyID := c.Param("lobby_id")

	finishedLobby, err := h.lobbyClient.FinishLobby(c.Request.Context(), lobbyID)
	if err != nil {
		var apiErr *gateway.APIError
		if errors.As(err, &apiErr) {
			c.JSON(apiErr.StatusCode, gin.H{"error": apiErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal error occurred"})
		return
	}

	c.JSON(http.StatusOK, finishedLobby)
}
