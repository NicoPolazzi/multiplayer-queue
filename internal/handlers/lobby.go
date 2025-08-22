package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
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

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.HTML(http.StatusInternalServerError, errorHTMLFilename, gin.H{"message": "Failed to read lobby creation response."})
			return
		}

		var newLobby lobby.Lobby
		if err := protojson.Unmarshal(body, &newLobby); err != nil {
			c.HTML(http.StatusInternalServerError, errorHTMLFilename, gin.H{"message": "Failed to parse lobby creation response."})
			return
		}

		c.Redirect(http.StatusSeeOther, "/lobbies/"+newLobby.LobbyId)
	} else {
		body, _ := io.ReadAll(resp.Body)
		c.HTML(http.StatusInternalServerError, errorHTMLFilename, gin.H{
			"message": fmt.Sprintf("Failed to create lobby: %s", string(body))})
	}
}

func (h *LobbyHandler) JoinLobby(c *gin.Context) {
	lobbyID := c.Param("lobby_id")
	usernameValue, _ := c.Get("username")
	username := usernameValue.(string)

	joinReq := &lobby.JoinLobbyRequest{
		LobbyId:  lobbyID,
		Username: username,
	}
	reqBody, _ := protojson.Marshal(joinReq)

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, h.gatewayBaseURL+"/api/v1/lobbies/"+lobbyID+"/join", bytes.NewBuffer(reqBody))
	if err != nil {
		c.HTML(http.StatusInternalServerError, errorHTMLFilename, gin.H{"message": "Failed to create join request."})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		c.HTML(http.StatusInternalServerError, errorHTMLFilename, gin.H{"message": "Failed to send join request to lobby service."})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		c.Redirect(http.StatusSeeOther, "/lobbies/"+lobbyID)
	} else {
		body, _ := io.ReadAll(resp.Body)
		c.HTML(http.StatusInternalServerError, errorHTMLFilename, gin.H{
			"message": fmt.Sprintf("Failed to join lobby: %s", string(body))})
	}
}

func (h *LobbyHandler) GetLobbyPage(c *gin.Context) {
	lobbyID := c.Param("lobby_id")

	resp, err := http.Get(h.gatewayBaseURL + "/api/v1/lobbies/" + lobbyID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, errorHTMLFilename, gin.H{"message": "Failed to get lobby details."})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.HTML(resp.StatusCode, errorHTMLFilename, gin.H{"message": fmt.Sprintf("Failed to get lobby details: %s", string(body))})
		return
	}

	var lobbyResponse lobby.Lobby
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.HTML(http.StatusInternalServerError, errorHTMLFilename, gin.H{"message": "Failed to read lobby details response."})
		return
	}
	if err := protojson.Unmarshal(body, &lobbyResponse); err != nil {
		c.HTML(http.StatusInternalServerError, errorHTMLFilename, gin.H{"message": "Failed to parse lobby details."})
		return
	}

	c.HTML(http.StatusOK, "lobby.html", gin.H{
		"lobby": &lobbyResponse,
	})
}

func (h *LobbyHandler) FinishLobby(c *gin.Context) {
	lobbyID := c.Param("lobby_id")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Forward the request to the gateway
	req, err := http.NewRequest(http.MethodPut, h.gatewayBaseURL+"/api/v1/lobbies/"+lobbyID+"/finish", bytes.NewBuffer(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request to gateway"})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to contact gateway"})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read gateway response"})
		return
	}

	c.Data(resp.StatusCode, "application/json", respBody)
}
