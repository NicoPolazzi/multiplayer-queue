package main

import (
	"github.com/NicoPolazzi/multiplayer-queue/gen/auth"
	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
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

// AppContainer holds all the dependencies for the application.
type AppContainer struct {
	RoutesManager *routes.RoutesManager
	LobbyService  lobby.LobbyServiceServer
	AuthService   auth.AuthServiceServer
}

func BuildContainer(db *gorm.DB, cfg *Config) *AppContainer {
	gatewayEndpoint := "http://" + cfg.Host + ":8081"

	// Repositories
	userRepo := usrRepo.NewSQLUserRepository(db)
	lobbyRepo := lobbyrepo.NewSQLLobbyRepository(db)

	// Services & Managers
	tokenManager := token.NewJWTTokenManager(cfg.JWTKey)

	// Handlers & Middleware
	userHandler := handlers.NewUserHandler(gatewayEndpoint)
	lobbyHandler := handlers.NewLobbyHandler(gatewayEndpoint)
	lobbyMiddleware := middleware.NewLobbyMiddleware(gatewayEndpoint)
	authMiddleware := middleware.NewAuthMiddleware(tokenManager)

	// Routes
	routesManager := routes.NewRoutes(userHandler, lobbyHandler, authMiddleware, lobbyMiddleware)

	// gRPC Services
	lobbyService := grpclobby.NewLobbyService(lobbyRepo, userRepo)
	authService := grpcauth.NewAuthService(userRepo, tokenManager)

	return &AppContainer{
		RoutesManager: routesManager,
		LobbyService:  lobbyService,
		AuthService:   authService,
	}
}
