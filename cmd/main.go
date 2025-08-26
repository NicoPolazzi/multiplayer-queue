package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/NicoPolazzi/multiplayer-queue/gen/auth"
	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// The program sets an http server using GIN to handle user requests. There requests are then translated by a gRPC
// gateway in RPCs handled by a gRPC server. Two gRPC services are available: auth and lobby. The former handles user
// authentication tasks, such as login and registration. The latter handles all lobby related task, such as the
// creation and the joining in a lobby.
func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := NewDatabaseConnection(cfg)
	if err != nil {
		log.Fatal(err)
	}

	container := BuildContainer(db, cfg)

	// Use an error group to manage concurrent servers
	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		return runGRPCServer(container, cfg)
	})

	g.Go(func() error {
		return runGRPCGateway(ctx, cfg)
	})

	g.Go(func() error {
		return runGinServer(container, cfg)
	})

	log.Println("Application started. Waiting for servers to exit.")
	if err := g.Wait(); err != nil {
		log.Fatalf("A server failed to run: %v", err)
	}
}

func runGRPCServer(container *AppContainer, cfg *Config) error {
	listenAddr := fmt.Sprintf(":%s", cfg.GRPCServerPort)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen for gRPC: %w", err)
	}
	s := grpc.NewServer()
	lobby.RegisterLobbyServiceServer(s, container.LobbyService)
	auth.RegisterAuthServiceServer(s, container.AuthService)

	log.Println("gRPC server listening at", lis.Addr())
	return s.Serve(lis)
}

func runGRPCGateway(ctx context.Context, cfg *Config) error {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("%s:%s", cfg.Host, cfg.GRPCServerPort)

	if err := lobby.RegisterLobbyServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return fmt.Errorf("failed to register Lobby gRPC gateway: %w", err)
	}
	if err := auth.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return fmt.Errorf("failed to register Auth gRPC gateway: %w", err)
	}

	listenAddr := fmt.Sprintf(":%s", cfg.GRPCGatewayPort)
	log.Println("gRPC gateway listening at", listenAddr)
	return http.ListenAndServe(listenAddr, mux)
}

func runGinServer(container *AppContainer, cfg *Config) error {
	gin.SetMode(cfg.GinMode)
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.LoadHTMLGlob("web/templates/*")
	container.RoutesManager.InitializeRoutes(router)

	listenAddr := fmt.Sprintf(":%s", cfg.GinServerPort)
	log.Printf("Gin Server is running on http://%s%s", cfg.Host, listenAddr)
	return router.Run(listenAddr)
}
