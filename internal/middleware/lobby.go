package middleware

import (
	"io"
	"log"
	"net/http"

	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
)

func LoadLobbies() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("lobbies", nil)

		isLoggedIn, exists := c.Get("is_logged_in")
		if !exists || !isLoggedIn.(bool) {
			c.Next()
			return
		}

		resp, err := http.Get("http://localhost:8081" + "/api/v1/lobbies")
		if err != nil {
			log.Printf("LobbyMiddleware: Could not connect to lobby service: %v", err)
			c.Next()
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("LobbyMiddleware: Failed to read response body: %v", err)
				c.Next()
				return
			}

			var lobbyList lobby.ListAvailableLobbiesResponse
			if err := protojson.Unmarshal(body, &lobbyList); err == nil {
				c.Set("lobbies", lobbyList.Lobbies)
			} else {
				log.Printf("LobbyMiddleware: Failed to parse lobby list: %v", err)
			}
		}

		c.Next()
	}
}
