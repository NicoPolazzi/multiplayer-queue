package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/encoding/protojson"
)

type LobbyMiddlewareTestSuite struct {
	suite.Suite
	middleware   *LobbyMiddleware
	recorder     *httptest.ResponseRecorder
	engine       *gin.Engine
	context      *gin.Context
	mockGateway  *httptest.Server
	mockResponse string
	mockStatus   int
}

func (s *LobbyMiddlewareTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.mockGateway = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(s.mockStatus)
		if _, err := fmt.Fprint(w, s.mockResponse); err != nil {
			s.FailNowf("Failed to write mock response", "Error: %v", err)
		}
	}))

	s.middleware = NewLobbyMiddleware(s.mockGateway.URL)
	s.recorder = httptest.NewRecorder()
	s.context, s.engine = gin.CreateTestContext(s.recorder)
}

func (s *LobbyMiddlewareTestSuite) TearDownTest() {
	s.mockGateway.Close()
}

func (s *LobbyMiddlewareTestSuite) TestLoadLobbiesSuccess() {
	s.mockStatus = http.StatusOK
	expectedLobbies := []*lobby.Lobby{
		{LobbyId: "lobby-1", Name: "Test Lobby 1"},
	}
	respProto := &lobby.ListAvailableLobbiesResponse{Lobbies: expectedLobbies}
	respJSON, _ := protojson.Marshal(respProto)
	s.mockResponse = string(respJSON)
	s.context.Set("is_logged_in", true)

	handler := s.middleware.LoadLobbies()
	handler(s.context)

	lobbiesValue, exists := s.context.Get("lobbies")
	s.True(exists, "lobbies should exist in context")
	lobbies, ok := lobbiesValue.([]*lobby.Lobby)
	s.True(ok, "lobbies should be of type []*lobby.Lobby")
	s.Len(lobbies, 1)
	s.Equal("lobby-1", lobbies[0].LobbyId)
}

func (s *LobbyMiddlewareTestSuite) TestLoadLobbiesWhenNotLoggedIn() {
	s.context.Set("is_logged_in", false)
	handler := s.middleware.LoadLobbies()
	handler(s.context)
	lobbiesValue, _ := s.context.Get("lobbies")
	s.Empty(lobbiesValue)
}

func (s *LobbyMiddlewareTestSuite) TestLoadLobbiesWhenGatewayIsDown() {
	s.mockGateway.Close()
	s.context.Set("is_logged_in", true)
	handler := s.middleware.LoadLobbies()
	handler(s.context)
	lobbiesValue, _ := s.context.Get("lobbies")
	errorTitle, _ := s.context.Get("ErrorTitle")
	errorMessage, _ := s.context.Get("ErrorMessage")
	s.Empty(lobbiesValue)
	s.Equal(lobbyErrorTitle, errorTitle)
	s.Equal("Could not retrieve the list of available lobbies. Please try again later.", errorMessage)
}

func (s *LobbyMiddlewareTestSuite) TestLoadLobbiesWhenGatewayReturnsError() {
	s.mockStatus = http.StatusInternalServerError
	s.context.Set("is_logged_in", true)
	handler := s.middleware.LoadLobbies()
	handler(s.context)
	errorTitle, _ := s.context.Get("ErrorTitle")
	errorMessage, _ := s.context.Get("ErrorMessage")
	lobbiesValue, _ := s.context.Get("lobbies")
	s.Empty(lobbiesValue)
	s.Equal(lobbyErrorTitle, errorTitle)
	s.Equal("There was a problem retrieving the list of lobbies.", errorMessage)
}

func (s *LobbyMiddlewareTestSuite) TestLoadLobbiesWhenResponseIsMalformed() {
	s.mockStatus = http.StatusOK
	s.mockResponse = `this is not valid json`
	s.context.Set("is_logged_in", true)
	handler := s.middleware.LoadLobbies()
	handler(s.context)
	lobbiesValue, _ := s.context.Get("lobbies")
	errorTitle, _ := s.context.Get("ErrorTitle")
	errorMessage, _ := s.context.Get("ErrorMessage")
	s.Empty(lobbiesValue)
	s.Equal(lobbyErrorTitle, errorTitle)
	s.Equal("Received an invalid response while fetching lobbies.", errorMessage)
}

func TestLobbyMiddleware(t *testing.T) {
	suite.Run(t, new(LobbyMiddlewareTestSuite))
}
