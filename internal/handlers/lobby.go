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

func CreateLobby(c *gin.Context) {
	lobbyName := c.PostForm("name")
	usernameValue, _ := c.Get("username")
	username := usernameValue.(string)

	createReq := &lobby.CreateLobbyRequest{
		Name:     lobbyName,
		Username: username,
	}
	reqBody, _ := protojson.Marshal(createReq)

	resp, err := http.Post("http://localhost:8081"+"/api/v1/lobbies", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"message": "Failed to send create request to lobby service."})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		c.Redirect(http.StatusSeeOther, "/")
	} else {
		body, _ := io.ReadAll(resp.Body)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"message": fmt.Sprintf("Failed to create lobby: %s", string(body))})
	}

}
