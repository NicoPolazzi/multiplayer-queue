package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type LobbyHandlerTestSuite struct {
	suite.Suite
	handler      *LobbyHandler
	recorder     *httptest.ResponseRecorder
	engine       *gin.Engine
	mockGateway  *httptest.Server
	mockResponse string
	mockStatus   int
}

func (s *LobbyHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)

	// Mock server to simulate the gateway API
	s.mockGateway = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(s.mockStatus)
		if _, err := fmt.Fprint(w, s.mockResponse); err != nil {
			s.FailNowf("Failed to write mock response", "Error: %v", err)
		}
	}))

	s.handler = NewLobbyHandler(s.mockGateway.URL)
	s.recorder = httptest.NewRecorder()
	_, s.engine = gin.CreateTestContext(s.recorder)
	s.engine.LoadHTMLGlob("../../web/templates/*")

	// This middleware mock is needed because the real authentication middleware isn't running in tests.
	// It ensures the "username" is available in the context for the handlers.
	s.engine.Use(func(c *gin.Context) {
		c.Set("is_logged_in", true)
		c.Set("username", "testuser")
		c.Next()
	})

	s.engine.POST("/lobbies/create", s.handler.CreateLobby)
	s.engine.POST("/lobbies/:lobby_id/join", s.handler.JoinLobby)
	s.engine.GET("/lobbies/:lobby_id", s.handler.GetLobbyPage)
	s.engine.PUT("/lobbies/:lobby_id/finish", s.handler.FinishLobby)
}

func (s *LobbyHandlerTestSuite) TearDownTest() {
	s.mockGateway.Close()
}

func (s *LobbyHandlerTestSuite) TestCreateLobbySuccess() {
	s.mockStatus = http.StatusOK
	s.mockResponse = `{"lobbyId": "lobby-123-abc"}`

	s.performPostRequest("/lobbies/create", "name=MyNewLobby")

	s.Equal(http.StatusSeeOther, s.recorder.Code, "Expected a redirect status")
	s.Equal("/lobbies/lobby-123-abc", s.recorder.Header().Get("Location"), "Expected redirect to the new lobby page")
}

func (s *LobbyHandlerTestSuite) performPostRequest(path, form string) {
	req, _ := http.NewRequest(http.MethodPost, path, strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	s.engine.ServeHTTP(s.recorder, req)
}

func (s *LobbyHandlerTestSuite) TestCreateLobbyWithEmptyName() {
	s.performPostRequest("/lobbies/create", "name=")

	s.Equal(http.StatusBadRequest, s.recorder.Code)
	s.assertErrorHTML("Lobby Creation Failed", "Lobby name cannot be empty.")
}

func (s *LobbyHandlerTestSuite) assertErrorHTML(expectedTitle, expectedMessage string) {
	body := s.recorder.Body.String()
	s.Contains(body, fmt.Sprintf("<strong>%s</strong>", expectedTitle), "Expected error title was not found in the response")
	s.Contains(body, fmt.Sprintf("<p>%s</p>", expectedMessage), "Expected error message was not found in the response")
}

func (s *LobbyHandlerTestSuite) TestCreateLobbyWhenGatewayFails() {
	s.mockStatus = http.StatusInternalServerError
	s.mockResponse = `{}`

	s.performPostRequest("/lobbies/create", "name=MyNewLobby")

	s.Equal(http.StatusInternalServerError, s.recorder.Code)
	s.assertErrorHTML("Lobby Creation Failed", "An unexpected error occurred while creating the lobby.")
}

func (s *LobbyHandlerTestSuite) TestCreateLobbyWhenGatewayIsDown() {
	s.mockGateway.Close() // Simulate the gateway being offline

	s.performPostRequest("/lobbies/create", "name=MyNewLobby")

	s.Equal(http.StatusInternalServerError, s.recorder.Code)
	s.assertErrorHTML("Lobby Creation Failed", "The server is currently unavailable. Please try again later.")
}

func (s *LobbyHandlerTestSuite) TestCreateLobbyFailsOnUnmarshal() {
	s.mockStatus = http.StatusOK
	s.mockResponse = `{"lobbyId": "123", "name": "Test"` // Malformed JSON (missing closing brace)

	s.performPostRequest("/lobbies/create", "name=MyNewLobby")

	s.Equal(http.StatusInternalServerError, s.recorder.Code)
	s.assertErrorHTML("Lobby Creation Failed", "Could not understand the server response.")
}

func (s *LobbyHandlerTestSuite) TestJoinLobbySuccess() {
	s.mockStatus = http.StatusOK
	s.mockResponse = `{}`

	s.performPostRequest("/lobbies/lobby-456/join", "")

	s.Equal(http.StatusSeeOther, s.recorder.Code)
	s.Equal("/lobbies/lobby-456", s.recorder.Header().Get("Location"))
}

func (s *LobbyHandlerTestSuite) TestJoinLobbyWhenGatewayFails() {
	s.mockStatus = http.StatusConflict
	s.mockResponse = `{"error": "lobby is full"}`

	s.performPostRequest("/lobbies/lobby-456/join", "")

	s.Equal(http.StatusInternalServerError, s.recorder.Code)
	s.assertErrorHTML("Join Lobby Failed", "An unexpected error occurred while joining the lobby.")
}

func (s *LobbyHandlerTestSuite) TestJoinLobbyWhenGatewayIsDown() {
	s.mockGateway.Close()

	s.performPostRequest("/lobbies/lobby-456/join", "")

	s.Equal(http.StatusInternalServerError, s.recorder.Code)
	s.assertErrorHTML("Join Lobby Failed", "The server is currently unavailable. Please try again later.")
}

func (s *LobbyHandlerTestSuite) TestGetLobbyPageSuccess() {
	s.mockStatus = http.StatusOK
	s.mockResponse = `{"lobbyId": "lobby-789", "name": "Test Lobby Name"}`

	s.performGetRequest("/lobbies/lobby-789")

	s.Equal(http.StatusOK, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), "Test Lobby Name")
}

func (s *LobbyHandlerTestSuite) performGetRequest(path string) {
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	s.engine.ServeHTTP(s.recorder, req)
}

func (s *LobbyHandlerTestSuite) TestFinishLobbySuccess() {
	s.mockStatus = http.StatusOK
	s.mockResponse = `{"message": "lobby finished", "winner": "player1"}`

	s.performPutRequest("/lobbies/lobby-abc/finish", `{"winnerId": "player1"}`)

	s.Equal(http.StatusOK, s.recorder.Code)
	s.JSONEq(s.mockResponse, s.recorder.Body.String())
	s.Equal("application/json", s.recorder.Header().Get("Content-Type"))
}

func (s *LobbyHandlerTestSuite) performPutRequest(path, jsonBody string) {
	req, _ := http.NewRequest(http.MethodPut, path, strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	s.engine.ServeHTTP(s.recorder, req)
}

func (s *LobbyHandlerTestSuite) TestFinishLobbyGatewayError() {
	s.mockStatus = http.StatusBadRequest
	s.mockResponse = `{"error": "winner not in lobby"}`

	s.performPutRequest("/lobbies/lobby-abc/finish", `{"winnerId": "player-not-in-lobby"}`)

	s.Equal(http.StatusBadRequest, s.recorder.Code)
	s.JSONEq(s.mockResponse, s.recorder.Body.String())
}

func TestLobbyHandlers(t *testing.T) {
	suite.Run(t, new(LobbyHandlerTestSuite))
}
