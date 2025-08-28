package main

import (
	"fmt"

	"github.com/NicoPolazzi/multiplayer-queue/gen/auth"
	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/NicoPolazzi/multiplayer-queue/internal/gateway"
	grpcauth "github.com/NicoPolazzi/multiplayer-queue/internal/grpc/auth"
	grpclobby "github.com/NicoPolazzi/multiplayer-queue/internal/grpc/lobby"
	"github.com/NicoPolazzi/multiplayer-queue/internal/handlers"
	"github.com/NicoPolazzi/multiplayer-queue/internal/middleware"
	lobbyrepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/lobby"
	usrRepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/user"
	"github.com/NicoPolazzi/multiplayer-queue/internal/routes"
	"github.com/NicoPolazzi/multiplayer-queue/internal/token"
	"gorm.io/gorm"
)

// AppContainer holds all the dependencies useful for the application.
type AppContainer struct {
	RoutesManager *routes.RoutesManager
	LobbyService  lobby.LobbyServiceServer
	AuthService   auth.AuthServiceServer
}

// BuildContainer is responsible to inject all the dependencies needed by the application.
func BuildContainer(db *gorm.DB, cfg *Config) *AppContainer {
	userRepo := usrRepo.NewSQLUserRepository(db)
	lobbyRepo := lobbyrepo.NewSQLLobbyRepository(db)

	tokenManager := token.NewJWTTokenManager([]byte(cfg.JWTSecret))

	gatewayURL := fmt.Sprintf("http://%s:%s", cfg.Host, cfg.GRPCGatewayPort)
	lobbyClient := gateway.NewLobbyGatewayClient(gatewayURL)
	userHandler := handlers.NewUserHandler(gatewayURL)
	lobbyHandler := handlers.NewLobbyHandler(lobbyClient)
	lobbyMiddleware := middleware.NewLobbyMiddleware(gatewayURL)
	authMiddleware := middleware.NewAuthMiddleware(tokenManager)

	routesManager := routes.NewRoutes(userHandler, lobbyHandler, authMiddleware, lobbyMiddleware)

	lobbyService := grpclobby.NewLobbyService(lobbyRepo, userRepo)
	authService := grpcauth.NewAuthService(userRepo, tokenManager)

	return &AppContainer{
		RoutesManager: routesManager,
		LobbyService:  lobbyService,
		AuthService:   authService,
	}
}
