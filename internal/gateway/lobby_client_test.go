package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestLobbyGatewayClientCreateLobby(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockResponse := &lobby.Lobby{LobbyId: "new-lobby-123", Name: "Test Lobby"}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			body, _ := protojson.Marshal(mockResponse)
			w.Write(body)
		}))
		defer server.Close()

		client := NewLobbyGatewayClient(server.URL)
		req := &lobby.CreateLobbyRequest{Name: "Test Lobby", Username: "creator"}
		res, err := client.CreateLobby(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, "new-lobby-123", res.LobbyId)
	})

	t.Run("Failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewLobbyGatewayClient(server.URL)
		req := &lobby.CreateLobbyRequest{Name: "Test Lobby", Username: "creator"}
		_, err := client.CreateLobby(context.Background(), req)

		require.Error(t, err)
		apiErr, ok := err.(*APIError)
		require.True(t, ok)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	})
}

func TestLobbyGatewayClientJoinLobby(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewLobbyGatewayClient(server.URL)
		req := &lobby.JoinLobbyRequest{LobbyId: "lobby-abc", Username: "player2"}
		err := client.JoinLobby(context.Background(), req)
		require.NoError(t, err)
	})

	t.Run("Failure - Lobby Full", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusConflict)
		}))
		defer server.Close()

		client := NewLobbyGatewayClient(server.URL)
		req := &lobby.JoinLobbyRequest{LobbyId: "lobby-abc", Username: "player3"}
		err := client.JoinLobby(context.Background(), req)

		require.Error(t, err)
		apiErr, ok := err.(*APIError)
		require.True(t, ok)
		assert.Equal(t, http.StatusConflict, apiErr.StatusCode)
	})
}

func TestLobbyGatewayClientFinishLobby(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		winnerId := uint32(1)
		mockResponse := &lobby.Lobby{LobbyId: "lobby-xyz", Status: "Finished", WinnerId: &winnerId}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			body, _ := protojson.Marshal(mockResponse)
			w.Write(body)
		}))
		defer server.Close()

		client := NewLobbyGatewayClient(server.URL)
		res, err := client.FinishLobby(context.Background(), "lobby-xyz")

		require.NoError(t, err)
		assert.Equal(t, "Finished", res.Status)
		assert.Equal(t, uint32(1), *res.WinnerId)
	})

	t.Run("Failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewLobbyGatewayClient(server.URL)
		_, err := client.FinishLobby(context.Background(), "lobby-xyz")

		require.Error(t, err)
		apiErr, ok := err.(*APIError)
		require.True(t, ok)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	})
}

func TestLobbyGatewayClientGetLobby(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockResponse := &lobby.Lobby{LobbyId: "lobby-abc", Name: "Existing Lobby"}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			body, _ := protojson.Marshal(mockResponse)
			w.Write(body)
		}))
		defer server.Close()

		client := NewLobbyGatewayClient(server.URL)
		res, err := client.GetLobby(context.Background(), "lobby-abc")

		require.NoError(t, err)
		assert.Equal(t, "lobby-abc", res.LobbyId)
	})

	t.Run("Failure - Not Found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := NewLobbyGatewayClient(server.URL)
		_, err := client.GetLobby(context.Background(), "non-existent-lobby")

		require.Error(t, err)
		apiErr, ok := err.(*APIError)
		require.True(t, ok)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	})
}

func TestLobbyGatewayClientListAvailableLobbies(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockResponse := &lobby.ListAvailableLobbiesResponse{
			Lobbies: []*lobby.Lobby{
				{LobbyId: "lobby-1", Name: "Lobby One"},
				{LobbyId: "lobby-2", Name: "Lobby Two"},
			},
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			body, _ := protojson.Marshal(mockResponse)
			w.Write(body)
		}))
		defer server.Close()

		client := NewLobbyGatewayClient(server.URL)
		lobbies, err := client.ListAvailableLobbies(context.Background())

		require.NoError(t, err)
		assert.Len(t, lobbies, 2)
		assert.Equal(t, "Lobby One", lobbies[0].Name)
	})

	t.Run("Failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewLobbyGatewayClient(server.URL)
		_, err := client.ListAvailableLobbies(context.Background())

		require.Error(t, err)
		apiErr, ok := err.(*APIError)
		require.True(t, ok)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	})
}
