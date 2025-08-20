package middleware

import (
	"io"
	"log"
	"net/http"

	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
)

type LobbyMiddleware struct {
	gatewayBaseURL string
}

const (
	lobbyErrorTitle = "Lobby Service Error"
)

func NewLobbyMiddleware(gatewayBaseURL string) *LobbyMiddleware {
	return &LobbyMiddleware{gatewayBaseURL: gatewayBaseURL}
}

// LoadLobbies calls the gateway to show the logged user the available lobbies' list.
// It hides the presence of a gateway from the user.
func (m *LobbyMiddleware) LoadLobbies() gin.HandlerFunc {
	return func(c *gin.Context) {
		if userIsLoggedIn(c) {
			resp := m.requestLobbiesToGateway(c)
			if resp != nil {
				defer func() {
					if err := resp.Body.Close(); err != nil {
						log.Printf("Error closing response body: %v", err)
					}
				}()

				if resp.StatusCode == http.StatusOK {
					fetchAndSetLobbies(resp, c)
				} else {
					setErrorValues(c, "There was a problem retrieving the list of lobbies.")
					log.Printf("LobbyMiddleware: Gateway returned non-OK status: %d", resp.StatusCode)
				}
			}
		}
		c.Next()
	}
}

func (m *LobbyMiddleware) requestLobbiesToGateway(c *gin.Context) *http.Response {
	resp, err := http.Get(m.gatewayBaseURL + "/api/v1/lobbies")
	if err != nil {
		setErrorValues(c, "Could not retrieve the list of available lobbies. Please try again later.")
		log.Printf("LobbyMiddleware: Could not connect to lobby service: %v", err)
	}

	return resp
}

func fetchAndSetLobbies(resp *http.Response, c *gin.Context) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		setErrorValues(c, "Received an unreadable response while fetching lobbies.")
		log.Printf("LobbyMiddleware: Failed to read response body: %v", err)
		return
	}

	var lobbyList lobby.ListAvailableLobbiesResponse
	if err := protojson.Unmarshal(body, &lobbyList); err == nil {
		c.Set("lobbies", lobbyList.Lobbies)
	} else {
		setErrorValues(c, "Received an invalid response while fetching lobbies.")
		log.Printf("LobbyMiddleware: Failed to parse lobby list: %v", err)
	}
}

func userIsLoggedIn(c *gin.Context) bool {
	isLoggedIn, exists := c.Get("is_logged_in")
	return exists && isLoggedIn.(bool)
}

func setErrorValues(c *gin.Context, message string) {
	c.Set("ErrorTitle", lobbyErrorTitle)
	c.Set("ErrorMessage", message)
}
