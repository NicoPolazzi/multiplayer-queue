package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/NicoPolazzi/multiplayer-queue/internal/gateway"
	"github.com/NicoPolazzi/multiplayer-queue/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/encoding/protojson"
)

type LobbyHandlerTestSuite struct {
	suite.Suite
	router      *gin.Engine
	mockGateway *httptest.Server
	handler     *LobbyHandler
	lobbyClient *gateway.LobbyGatewayClient
}

func (s *LobbyHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.router = gin.Default()
	s.router.LoadHTMLGlob("../../web/templates/*")
}

func (s *LobbyHandlerTestSuite) AfterTest() {
	if s.mockGateway != nil {
		s.mockGateway.Close()
	}
}

func (s *LobbyHandlerTestSuite) setup(mockHandler http.HandlerFunc) {
	s.mockGateway = httptest.NewServer(mockHandler)
	s.lobbyClient = gateway.NewLobbyGatewayClient(s.mockGateway.URL)
	s.handler = NewLobbyHandler(s.lobbyClient)

	// Add a mock middleware to simulate a logged-in user.
	s.router.Use(func(c *gin.Context) {
		middleware.SetUserInContext(c, &middleware.User{Username: "testuser"})
		c.Next()
	})
}

func (s *LobbyHandlerTestSuite) TestCreateLobbySuccess() {
	s.setup(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := &lobby.Lobby{LobbyId: "lobby-123"}
		body, _ := protojson.Marshal(resp)
		_, err := w.Write(body)
		if err != nil {
			s.T().Fatalf("Failed to write response: %v", err)
		}
	})
	s.router.POST("/lobbies/create", s.handler.CreateLobby)

	formData := url.Values{"name": {"My New Lobby"}}
	req, _ := http.NewRequest(http.MethodPost, "/lobbies/create", strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusSeeOther, w.Code)
	s.Equal("/lobbies/lobby-123", w.Header().Get("Location"))
}

func (s *LobbyHandlerTestSuite) TestCreateLobbyFailsWithEmptyName() {
	s.setup(nil)
	s.router.POST("/lobbies/create", s.handler.CreateLobby)

	formData := url.Values{"name": {""}}
	req, _ := http.NewRequest(http.MethodPost, "/lobbies/create", strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
	s.Contains(w.Body.String(), "Lobby Creation Failed")
	s.Contains(w.Body.String(), "Lobby name cannot be empty.")
}

func (s *LobbyHandlerTestSuite) TestCreateLobbyFailsWhenTheGatewayFails() {
	s.setup(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	s.router.POST("/lobbies/create", s.handler.CreateLobby)

	formData := url.Values{"name": {"A Lobby"}}
	req, _ := http.NewRequest(http.MethodPost, "/lobbies/create", strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), "An unexpected error occurred")
}

func (s *LobbyHandlerTestSuite) TestJoinLobbySuccess() {
	s.setup(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := &lobby.Lobby{LobbyId: "lobby-456"}
		body, _ := protojson.Marshal(resp)
		_, err := w.Write(body)
		if err != nil {
			s.T().Fatalf("Failed to write response: %v", err)
		}
	})
	s.router.POST("/lobbies/:lobby_id/join", s.handler.JoinLobby)

	req, _ := http.NewRequest(http.MethodPost, "/lobbies/lobby-456/join", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusSeeOther, w.Code)
	s.Equal("/lobbies/lobby-456", w.Header().Get("Location"))
}

func (s *LobbyHandlerTestSuite) TestJoinLobbyGatewayFailure() {
	s.setup(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	s.router.POST("/lobbies/:lobby_id/join", s.handler.JoinLobby)

	req, _ := http.NewRequest(http.MethodPost, "/lobbies/any-id/join", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), "Join Lobby Failed")
	s.Contains(w.Body.String(), "An unexpected error occurred while joining the lobby.")
}

func (s *LobbyHandlerTestSuite) TestGetLobbyPageSuccess() {
	s.setup(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := &lobby.Lobby{LobbyId: "lobby-789", Name: "The Best Lobby"}
		body, _ := protojson.Marshal(resp)
		_, err := w.Write(body)
		if err != nil {
			s.T().Fatalf("Failed to write response: %v", err)
		}
	})
	s.router.GET("/lobbies/:lobby_id", s.handler.GetLobbyPage)

	req, _ := http.NewRequest(http.MethodGet, "/lobbies/lobby-789", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "The Best Lobby")
}

func (s *LobbyHandlerTestSuite) TestGetLobbyPageGatewayFailure() {
	s.setup(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	s.router.GET("/lobbies/:lobby_id", s.handler.GetLobbyPage)

	req, _ := http.NewRequest(http.MethodGet, "/lobbies/any-id", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), "Error Fetching Lobby")
	s.Contains(w.Body.String(), "The server is currently unavailable.")
}

func (s *LobbyHandlerTestSuite) TestGetLobbyPageNotFound() {
	s.setup(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	s.router.GET("/lobbies/:lobby_id", s.handler.GetLobbyPage)

	req, _ := http.NewRequest(http.MethodGet, "/lobbies/not-found-id", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
	s.Contains(w.Body.String(), "Lobby Not Found")
}

func (s *LobbyHandlerTestSuite) TestFinishLobbySuccess() {
	s.setup(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := &lobby.Lobby{LobbyId: "lobby-abc", Status: "Finished"}
		body, _ := protojson.Marshal(resp)
		_, err := w.Write(body)
		if err != nil {
			s.T().Fatalf("Failed to write response: %v", err)
		}
	})
	s.router.PUT("/lobbies/:lobby_id/finish", s.handler.FinishLobby)

	req, _ := http.NewRequest(http.MethodPut, "/lobbies/lobby-abc/finish", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.JSONEq(`{"lobby_id":"lobby-abc", "status":"Finished"}`, w.Body.String())
}

func (s *LobbyHandlerTestSuite) TestFinishLobbyGatewayFailure() {
	s.setup(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(`{"error": "database unavailable"}`))
		if err != nil {
			s.T().Fatalf("Failed to write response: %v", err)
		}
	})
	s.router.PUT("/lobbies/:lobby_id/finish", s.handler.FinishLobby)

	req, _ := http.NewRequest(http.MethodPut, "/lobbies/any-id/finish", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)
	s.JSONEq(`{"error": "An unexpected error occurred"}`, w.Body.String())
}

func TestLobbyHandler(t *testing.T) {
	suite.Run(t, new(LobbyHandlerTestSuite))
}
