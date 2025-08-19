package middleware

import (
	"io"
	"net/http"

	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
)

type LobbyMiddleware struct {
	gatewayBaseURL string
}

func NewLobbyMiddleware(gatewayBaseURL string) *LobbyMiddleware {
	return &LobbyMiddleware{gatewayBaseURL: gatewayBaseURL}
}

func (m *LobbyMiddleware) LoadLobbies() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("lobbies", nil)

		isLoggedIn, exists := c.Get("is_logged_in")
		if !exists || !isLoggedIn.(bool) {
			c.Next()
			return
		}

		resp, err := http.Get(m.gatewayBaseURL + "/api/v1/lobbies")
		if err != nil {
			c.Set("ErrorTitle", "Lobby Service Error")
			c.Set("ErrorMessage", "Could not retrieve the list of available lobbies. Please try again later.")
			c.Next()
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			var lobbyList lobby.ListAvailableLobbiesResponse
			if err := protojson.Unmarshal(body, &lobbyList); err == nil {
				c.Set("lobbies", lobbyList.Lobbies)
			} else {
				c.Set("ErrorTitle", "Lobby Service Error")
				c.Set("ErrorMessage", "Received an invalid response while fetching lobbies.")
			}
		}

		c.Next()
	}
}
