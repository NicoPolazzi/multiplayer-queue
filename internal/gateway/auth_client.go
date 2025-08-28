package gateway

import (
	"context"
	"net/http"

	"github.com/NicoPolazzi/multiplayer-queue/gen/auth"
)

type AuthGatewayClient struct {
	*baseClient
}

func NewAuthGatewayClient(baseURL string) *AuthGatewayClient {
	return &AuthGatewayClient{
		&baseClient{
			baseURL:    baseURL,
			httpClient: &http.Client{},
		},
	}
}

func (c *AuthGatewayClient) Login(ctx context.Context, req *auth.LoginUserRequest) (*auth.LoginUserResponse, error) {
	var loginResponse auth.LoginUserResponse
	err := c.doProtoRequest(ctx, http.MethodPost, "/api/v1/auth/login", req, &loginResponse)
	return &loginResponse, err
}

func (c *AuthGatewayClient) Register(ctx context.Context, req *auth.RegisterUserRequest) error {
	err := c.doProtoRequest(ctx, http.MethodPost, "/api/v1/auth/register", req, nil)
	return err
}
