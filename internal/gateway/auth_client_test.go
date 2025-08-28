package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/gen/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestAuthGatewayClientLogin(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockResponse := &auth.LoginUserResponse{
			Token: "mock-jwt-token",
			User:  &auth.User{Id: 1, Username: "testuser"},
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			body, _ := protojson.Marshal(mockResponse)
			_, err := w.Write(body)
			if err != nil {
				t.Fatalf("Failed to write response: %v", err)
			}
		}))
		defer server.Close()

		client := NewAuthGatewayClient(server.URL)
		req := &auth.LoginUserRequest{Username: "testuser", Password: "password"}

		res, err := client.Login(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "mock-jwt-token", res.Token)
		assert.Equal(t, "testuser", res.User.Username)
	})

	t.Run("Failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		client := NewAuthGatewayClient(server.URL)
		req := &auth.LoginUserRequest{Username: "testuser", Password: "wrong"}

		res, err := client.Login(context.Background(), req)
		require.Error(t, err)
		apiErr, ok := err.(*APIError)
		require.True(t, ok, "error should be of type APIError")
		assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
		assert.Nil(t, res.User)
	})
}

func TestAuthGatewayClientRegister(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewAuthGatewayClient(server.URL)
		req := &auth.RegisterUserRequest{Username: "newuser", Password: "password"}

		err := client.Register(context.Background(), req)
		require.NoError(t, err)
	})

	t.Run("Failure - Conflict", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusConflict)
		}))
		defer server.Close()

		client := NewAuthGatewayClient(server.URL)
		req := &auth.RegisterUserRequest{Username: "existinguser", Password: "password"}

		err := client.Register(context.Background(), req)
		require.Error(t, err)
		apiErr, ok := err.(*APIError)
		require.True(t, ok)
		assert.Equal(t, http.StatusConflict, apiErr.StatusCode)
	})
}
