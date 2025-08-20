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

// LoadLobbies calls the gateway to show the logged user the list of the available lobbies.
// It hides the presence of a gateway from the user.
func (m *LobbyMiddleware) LoadLobbies() gin.HandlerFunc {
	return func(c *gin.Context) {
		if userIsLoggedIn(c) {
			resp, err := http.Get(m.gatewayBaseURL + "/api/v1/lobbies")
			if err != nil {
				setErrorValues(c, "Lobby Service Error",
					"Could not retrieve the list of available lobbies. Please try again later.")
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
					setErrorValues(c, "Lobby Service Error", "Received an invalid response while fetching lobbies.")
				}
			}
		}

		c.Next()
	}
}

func userIsLoggedIn(c *gin.Context) bool {
	isLoggedIn, exists := c.Get("is_logged_in")
	return exists && isLoggedIn.(bool)
}

func setErrorValues(c *gin.Context, title, message string) {
	c.Set("ErrorTitle", title)
	c.Set("ErrorMessage", message)
}
