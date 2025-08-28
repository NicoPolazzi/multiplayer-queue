package gateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
)

type LobbyGatewayClient struct {
	*baseClient
}

func NewLobbyGatewayClient(baseURL string) *LobbyGatewayClient {
	return &LobbyGatewayClient{
		&baseClient{
			baseURL:    baseURL,
			httpClient: &http.Client{},
		},
	}
}

func (c *LobbyGatewayClient) CreateLobby(ctx context.Context, req *lobby.CreateLobbyRequest) (*lobby.Lobby, error) {
	var newLobby lobby.Lobby
	err := c.doProtoRequest(ctx, http.MethodPost, "/api/v1/lobbies", req, &newLobby)
	if err != nil {
		return nil, err
	}
	return &newLobby, nil
}

func (c *LobbyGatewayClient) JoinLobby(ctx context.Context, req *lobby.JoinLobbyRequest) error {
	path := fmt.Sprintf("/api/v1/lobbies/%s/join", req.LobbyId)
	return c.doProtoRequest(ctx, http.MethodPut, path, req, nil)
}

func (c *LobbyGatewayClient) GetLobby(ctx context.Context, lobbyID string) (*lobby.Lobby, error) {
	var foundLobby lobby.Lobby
	path := fmt.Sprintf("/api/v1/lobbies/%s", lobbyID)
	err := c.doProtoRequest(ctx, http.MethodGet, path, nil, &foundLobby)
	if err != nil {
		return nil, err
	}
	return &foundLobby, nil
}

func (c *LobbyGatewayClient) FinishLobby(ctx context.Context, lobbyID string) (*lobby.Lobby, error) {
	var finishedLobby lobby.Lobby
	path := fmt.Sprintf("/api/v1/lobbies/%s/finish", lobbyID)
	err := c.doProtoRequest(ctx, http.MethodPut, path, nil, &finishedLobby)
	if err != nil {
		return nil, err
	}
	return &finishedLobby, nil
}

func (c *LobbyGatewayClient) ListAvailableLobbies(ctx context.Context) ([]*lobby.Lobby, error) {
	var lobbyListResponse lobby.ListAvailableLobbiesResponse
	err := c.doProtoRequest(ctx, http.MethodGet, "/api/v1/lobbies/available", nil, &lobbyListResponse)
	if err != nil {
		return nil, err
	}
	return lobbyListResponse.Lobbies, nil
}
