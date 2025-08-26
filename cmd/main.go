package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/NicoPolazzi/multiplayer-queue/gen/auth"
	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// The program sets an http server using GIN to handle user requests. There requests are then translated by a gRPC
// gateway in RPCs handled by a gRPC server. Two gRPC services are available: auth and lobby. The former handles user
// authentication tasks, such as login and registration. The latter handles all lobby related task, such as the
// creation and the joining in a lobby.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	db, err := NewDatabaseConnection(cfg)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	container := BuildContainer(db, cfg)

	var wg sync.WaitGroup
	errChan := make(chan error, 3)

	// Start the gRPC server.
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := runGRPCServer(ctx, container, cfg); err != nil {
			errChan <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	// Start the gRPC gateway.
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := runGRPCGateway(ctx, cfg); err != nil {
			errChan <- fmt.Errorf("gRPC gateway error: %w", err)
		}
	}()

	// Start the Gin web server.
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := runGinServer(ctx, container, cfg); err != nil {
			errChan <- fmt.Errorf("gin server error: %w", err)
		}
	}()

	log.Println("Application started. Press Ctrl+C to shut down.")

	select {
	case err := <-errChan:
		log.Printf("An unrecoverable server error occurred: %v", err)
		stop()
	case <-ctx.Done():
		log.Println("Shutdown signal received.")
	}

	// Wait for all server goroutines to complete their graceful shutdown.
	wg.Wait()
	log.Println("All servers have been shut down gracefully.")
}

func runGRPCServer(ctx context.Context, container *AppContainer, cfg *Config) error {
	listenAddr := fmt.Sprintf(":%s", cfg.GRPCServerPort)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen for gRPC on %s: %w", listenAddr, err)
	}

	s := grpc.NewServer()
	lobby.RegisterLobbyServiceServer(s, container.LobbyService)
	auth.RegisterAuthServiceServer(s, container.AuthService)

	go func() {
		<-ctx.Done()
		log.Println("Shutting down gRPC server...")
		s.GracefulStop()
	}()

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
	srv := &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		log.Println("Shutting down gRPC gateway...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error shutting down gRPC gateway: %v", err)
		}
	}()

	log.Println("gRPC gateway listening at", listenAddr)
	return srv.ListenAndServe()
}

func runGinServer(ctx context.Context, container *AppContainer, cfg *Config) error {
	gin.SetMode(cfg.GinMode)
	router := gin.Default()
	router.LoadHTMLGlob("web/templates/*")
	container.RoutesManager.InitializeRoutes(router)

	listenAddr := fmt.Sprintf(":%s", cfg.GinServerPort)
	srv := &http.Server{
		Addr:    listenAddr,
		Handler: router,
	}

	// Start a goroutine to listen for the context cancellation.
	go func() {
		<-ctx.Done()
		log.Println("Shutting down Gin server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error shutting down Gin server: %v", err)
		}
	}()

	log.Printf("Gin Server is running on http://%s%s", cfg.Host, listenAddr)
	return srv.ListenAndServe()
}
