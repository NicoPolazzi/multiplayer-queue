package middleware

import (
	"io"
	"log"
	"net/http"

	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	lobbyErrorTitle = "Lobby Service Error"
)

type LobbyMiddleware struct {
	gatewayBaseURL string
}

func NewLobbyMiddleware(gatewayBaseURL string) *LobbyMiddleware {
	return &LobbyMiddleware{gatewayBaseURL: gatewayBaseURL}
}

// LoadLobbies calls the gateway to show the logged user the available lobbies' list.
// It hides the presence of a gateway from the user.
func (m *LobbyMiddleware) LoadLobbies() gin.HandlerFunc {
	return func(c *gin.Context) {
		isLoggedIn := c.GetBool("is_logged_in")

		// Always set a default value to prevent panics in the template.
		c.Set("lobbies", []*lobby.Lobby{})

		// Only fetch lobbies if the user is logged in.
		if isLoggedIn {
			resp, err := http.Get(m.gatewayBaseURL + "/api/v1/lobbies/available")
			if err != nil {
				log.Printf("LobbyMiddleware: Could not connect to lobby service: %v", err)
				c.Set("ErrorTitle", "Service Error")
				c.Set("ErrorMessage", "Could not retrieve the list of available lobbies. Please try again later.")
				c.Next()
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("LobbyMiddleware: Gateway returned non-OK status: %d", resp.StatusCode)
				c.Set("ErrorTitle", "Service Error")
				c.Set("ErrorMessage", "There was a problem retrieving the list of lobbies.")
				c.Next()
				return
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("LobbyMiddleware: Failed to read response body: %v", err)
				c.Set("ErrorTitle", "Response Error")
				c.Set("ErrorMessage", "Received an unreadable response while fetching lobbies.")
				c.Next()
				return
			}

			var lobbyList lobby.ListAvailableLobbiesResponse
			if err := protojson.Unmarshal(body, &lobbyList); err != nil {
				log.Printf("LobbyMiddleware: Failed to parse lobby list: %v", err)
				c.Set("ErrorTitle", "Response Error")
				c.Set("ErrorMessage", "Received an invalid response while fetching lobbies.")
				c.Next()
				return
			}

			// On success, overwrite the default empty slice with the real data.
			c.Set("lobbies", lobbyList.Lobbies)
		}

		c.Next()
	}
}
