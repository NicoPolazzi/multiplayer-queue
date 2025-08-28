package gateway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// A more specific error to return when the API gives a non-200 response
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error: status %d, message: %s", e.StatusCode, e.Message)
}

type LobbyGatewayClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewLobbyGatewayClient(baseURL string) *LobbyGatewayClient {
	return &LobbyGatewayClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c *LobbyGatewayClient) CreateLobby(ctx context.Context, req *lobby.CreateLobbyRequest) (*lobby.Lobby, error) {
	var newLobby lobby.Lobby
	err := c.doProtoRequest(ctx, http.MethodPost, "/api/v1/lobbies", req, &newLobby)
	return &newLobby, err
}

func (c *LobbyGatewayClient) JoinLobby(ctx context.Context, req *lobby.JoinLobbyRequest) error {
	path := fmt.Sprintf("/api/v1/lobbies/%s/join", req.LobbyId)
	return c.doProtoRequest(ctx, http.MethodPut, path, req, nil)
}

func (c *LobbyGatewayClient) GetLobby(ctx context.Context, lobbyID string) (*lobby.Lobby, error) {
	var foundLobby lobby.Lobby
	path := fmt.Sprintf("/api/v1/lobbies/%s", lobbyID)
	err := c.doProtoRequest(ctx, http.MethodGet, path, nil, &foundLobby)
	return &foundLobby, err
}

func (c *LobbyGatewayClient) FinishLobby(ctx context.Context, lobbyID string) (*lobby.Lobby, error) {
	var finishedLobby lobby.Lobby
	path := fmt.Sprintf("/api/v1/lobbies/%s/finish", lobbyID)
	err := c.doProtoRequest(ctx, http.MethodPut, path, nil, &finishedLobby)
	return &finishedLobby, err
}

// doProtoRequest create and execute an HTTP request for the gateway.
// If the user specifies a res, the value is populated with the unmarshaled response.
func (c *LobbyGatewayClient) doProtoRequest(ctx context.Context, method, path string, req, res proto.Message) error {
	var reqBody io.Reader
	if req != nil {
		bodyBytes, err := protojson.Marshal(req)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(bodyBytes)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    "An unexpected error occurred",
		}
	}

	if res != nil {
		if err := protojson.Unmarshal(body, res); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}
