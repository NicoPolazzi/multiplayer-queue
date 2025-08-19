package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	CreateLobbyPath   = "/lobbies/create"
	errorHTMLFilename = "error.html"
)

type LobbyHandler struct {
	gatewayBaseURL string
}

func NewLobbyHandler(gatewayBaseURL string) *LobbyHandler {
	return &LobbyHandler{gatewayBaseURL: gatewayBaseURL}
}

func (h *LobbyHandler) CreateLobby(c *gin.Context) {
	lobbyName := c.PostForm("name")
	if lobbyName == "" {
		c.HTML(http.StatusBadRequest, errorHTMLFilename, gin.H{"message": "Lobby name cannot be empty."})
		return
	}

	usernameValue, _ := c.Get("username")
	username := usernameValue.(string)

	createReq := &lobby.CreateLobbyRequest{
		Name:     lobbyName,
		Username: username,
	}
	reqBody, _ := protojson.Marshal(createReq)

	resp, err := http.Post(h.gatewayBaseURL+"/api/v1/lobbies", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		c.HTML(http.StatusInternalServerError, errorHTMLFilename, gin.H{
			"message": "Failed to send create request to lobby service."})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		c.Redirect(http.StatusSeeOther, "/")
	} else {
		body, _ := io.ReadAll(resp.Body)
		c.HTML(http.StatusInternalServerError, errorHTMLFilename, gin.H{
			"message": fmt.Sprintf("Failed to create lobby: %s", string(body))})
	}

}
