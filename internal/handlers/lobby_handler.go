package handlers

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	indexPageFilename = "index.html"
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
		c.HTML(http.StatusBadRequest, indexPageFilename, gin.H{
			"ErrorTitle":   "Lobby Creation Failed",
			"ErrorMessage": "Lobby name cannot be empty.",
		})
		return
	}

	usernameValue, _ := c.Get("username")
	username := usernameValue.(string)

	createReq := &lobby.CreateLobbyRequest{
		Name:     lobbyName,
		Username: username,
	}
	reqBody, err := protojson.Marshal(createReq)
	if err != nil {
		c.HTML(http.StatusInternalServerError, indexPageFilename, gin.H{
			"ErrorTitle":   "Lobby Creation Failed",
			"ErrorMessage": "Could not prepare the request.",
		})
		return
	}

	resp, err := http.Post(h.gatewayBaseURL+"/api/v1/lobbies", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		c.HTML(http.StatusInternalServerError, indexPageFilename, gin.H{
			"ErrorTitle":   "Lobby Creation Failed",
			"ErrorMessage": "The server is currently unavailable. Please try again later.",
		})
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.HTML(http.StatusInternalServerError, indexPageFilename, gin.H{
			"ErrorTitle":   "Lobby Creation Failed",
			"ErrorMessage": "An unexpected error occurred while creating the lobby.",
			"is_logged_in": true,
			"username":     username,
		})
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.HTML(http.StatusInternalServerError, indexPageFilename, gin.H{
			"ErrorTitle":   "Lobby Creation Failed",
			"ErrorMessage": "Could not read the server response.",
		})
		return
	}

	var newLobby lobby.Lobby
	if err := protojson.Unmarshal(body, &newLobby); err != nil {
		c.HTML(http.StatusInternalServerError, indexPageFilename, gin.H{
			"ErrorTitle":   "Lobby Creation Failed",
			"ErrorMessage": "Could not understand the server response.",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/lobbies/"+newLobby.LobbyId)
}

func (h *LobbyHandler) JoinLobby(c *gin.Context) {
	lobbyID := c.Param("lobby_id")
	usernameValue, _ := c.Get("username")
	username := usernameValue.(string)

	joinReq := &lobby.JoinLobbyRequest{
		LobbyId:  lobbyID,
		Username: username,
	}
	reqBody, err := protojson.Marshal(joinReq)
	if err != nil {
		c.HTML(http.StatusInternalServerError, indexPageFilename, gin.H{
			"ErrorTitle":   "Join Lobby Failed",
			"ErrorMessage": "Could not prepare the request.",
		})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, h.gatewayBaseURL+"/api/v1/lobbies/"+lobbyID+"/join", bytes.NewBuffer(reqBody))
	if err != nil {
		c.HTML(http.StatusInternalServerError, indexPageFilename, gin.H{
			"ErrorTitle":   "Join Lobby Failed",
			"ErrorMessage": "Could not create the join request.",
		})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		c.HTML(http.StatusInternalServerError, indexPageFilename, gin.H{
			"ErrorTitle":   "Join Lobby Failed",
			"ErrorMessage": "The server is currently unavailable. Please try again later.",
		})
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode == http.StatusOK {
		c.Redirect(http.StatusSeeOther, "/lobbies/"+lobbyID)
	} else {
		c.HTML(http.StatusInternalServerError, indexPageFilename, gin.H{
			"ErrorTitle":   "Join Lobby Failed",
			"ErrorMessage": "An unexpected error occurred while joining the lobby.",
		})
	}
}

func (h *LobbyHandler) GetLobbyPage(c *gin.Context) {
	lobbyID := c.Param("lobby_id")

	resp, err := http.Get(h.gatewayBaseURL + "/api/v1/lobbies/" + lobbyID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, indexPageFilename, gin.H{
			"ErrorTitle":   "Error Fetching Lobby",
			"ErrorMessage": "The server is currently unavailable.",
		})
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.HTML(resp.StatusCode, "lobby.html", gin.H{
			"ErrorTitle":   "Error Fetching Lobby",
			"ErrorMessage": "Could not find the requested lobby.",
		})
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "lobby.html", gin.H{
			"ErrorTitle":   "Error Fetching Lobby",
			"ErrorMessage": "Failed to read the server's response.",
		})
		return
	}

	var lobbyResponse lobby.Lobby
	if err := protojson.Unmarshal(body, &lobbyResponse); err != nil {
		c.HTML(http.StatusInternalServerError, "lobby.html", gin.H{
			"ErrorTitle":   "Error Fetching Lobby",
			"ErrorMessage": "Failed to understand the server's response.",
		})
		return
	}

	c.HTML(http.StatusOK, "lobby.html", gin.H{
		"lobby":        &lobbyResponse,
		"is_logged_in": c.GetBool("is_logged_in"),
		"username":     c.GetString("username"),
	})
}

func (h *LobbyHandler) FinishLobby(c *gin.Context) {
	lobbyID := c.Param("lobby_id")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read gateway response"})
		return
	}

	// Proxy the status code and body from the gateway
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}
