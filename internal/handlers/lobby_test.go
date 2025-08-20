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

	s.mockGateway = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(s.mockStatus)
		if _, err := fmt.Fprint(w, s.mockResponse); err != nil {
			s.FailNowf("Failed to write mock response", "Error: %v", err)
		}
	}))
	s.handler = NewLobbyHandler(s.mockGateway.URL)
	s.recorder = httptest.NewRecorder()
	_, s.engine = gin.CreateTestContext(s.recorder)
	s.engine.LoadHTMLGlob("../../web/templates/*")

	// This middleware mock is needed because the real one isn't running.
	// It ensures the "username" is available in the context for the handlers.
	s.engine.Use(func(c *gin.Context) {
		c.Set("is_logged_in", true)
		c.Set("username", "testuser")
		c.Next()
	})

	s.engine.POST(CreateLobbyPath, s.handler.CreateLobby)
}

func (s *LobbyHandlerTestSuite) TearDownTest() {
	s.mockGateway.Close()
}

func (s *LobbyHandlerTestSuite) TestCreateLobbySuccess() {
	s.mockStatus = http.StatusOK
	s.mockResponse = `{"status": "ok"}`
	s.performPostRequest(CreateLobbyPath, "name=MyNewLobby")
	s.Equal(http.StatusSeeOther, s.recorder.Code)
	s.Equal("/", s.recorder.Header().Get("Location"))
}

func (s *LobbyHandlerTestSuite) performPostRequest(path, form string) {
	req, _ := http.NewRequest(http.MethodPost, path, strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	s.engine.ServeHTTP(s.recorder, req)
}

func (s *LobbyHandlerTestSuite) TestCreateLobbyWhenGatewayFails() {
	s.mockStatus = http.StatusInternalServerError
	s.mockResponse = `{"error": "gateway not working"}`
	s.performPostRequest(CreateLobbyPath, "name=MyNewLobby")
	s.Equal(http.StatusInternalServerError, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), "Failed to create lobby")
	s.Contains(s.recorder.Body.String(), "gateway not working")
}

func (s *LobbyHandlerTestSuite) TestCreateLobbyWithEmptyName() {
	s.performPostRequest(CreateLobbyPath, "name=")
	s.Equal(http.StatusBadRequest, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), "name cannot be empty")
}

func (s *LobbyHandlerTestSuite) TestCreateLobbyWhenGatewayIsDown() {
	s.mockGateway.Close()
	s.performPostRequest(CreateLobbyPath, "name=MyNewLobby")
	s.Equal(http.StatusInternalServerError, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), "Failed to send create request to lobby service.")
}

func TestLobbyHandlers(t *testing.T) {
	suite.Run(t, new(LobbyHandlerTestSuite))
}
