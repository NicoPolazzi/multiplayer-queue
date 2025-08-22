package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	grpcServer "github.com/NicoPolazzi/multiplayer-queue/internal/grpc"
	"github.com/NicoPolazzi/multiplayer-queue/internal/handlers"
	"github.com/NicoPolazzi/multiplayer-queue/internal/middleware"
	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	lobbyrepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/lobby"
	usrRepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/user"
	"github.com/NicoPolazzi/multiplayer-queue/internal/routes"
	"github.com/NicoPolazzi/multiplayer-queue/internal/service"
	"github.com/NicoPolazzi/multiplayer-queue/internal/token"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Check values in .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET not set in .env file")
	}
	key := []byte(jwtSecret)

	// Set db connection and schema
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.Lobby{}); err != nil {
		log.Fatal("migration failed:", err)
	}

	userRepo := usrRepo.NewSQLUserRepository(db)
	lobbyRepo := lobbyrepo.NewSQLLobbyRepository(db)
	tokenManager := token.NewJWTTokenManager(key)
	authService := service.NewJWTAuthService(userRepo, tokenManager)
	userHandler := handlers.NewUserHandler(authService)
	lobbyHandler := handlers.NewLobbyHandler("http://" + host + ":8081")
	lobbyMiddleware := middleware.NewLobbyMiddleware("http://" + host + ":8081")
	authMiddleware := middleware.NewAuthMiddleware(tokenManager)
	routesManager := routes.NewRoutes(userHandler, lobbyHandler, authMiddleware, lobbyMiddleware)

	// TODO: review the initialization of the server and gateway
	// gRPC server setup
	go func() {
		lis, err := net.Listen("tcp", host+":9090")
		if err != nil {
			log.Fatalf("failed to listen for gRPC: %v", err)
		}
		s := grpc.NewServer()
		lobbyServer := grpcServer.NewLobbyServer(lobbyRepo, userRepo)
		lobby.RegisterLobbyServiceServer(s, lobbyServer)
		log.Println("gRPC server listening at", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// gRPC gateway setup
	go func() {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		mux := runtime.NewServeMux()
		opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
		grpcServerEndpoint := host + ":9090"
		err := lobby.RegisterLobbyServiceHandlerFromEndpoint(ctx, mux, grpcServerEndpoint, opts)
		if err != nil {
			log.Fatalf("Failed to register gRPC gateway: %v", err)
		}

		log.Println("gRPC gateway listening at :8081")
		if err := http.ListenAndServe(":8081", mux); err != nil {
			log.Fatalf("Failed to serve gRPC gateway: %v", err)
		}
	}()

	router := gin.Default()
	router.LoadHTMLGlob("web/templates/*")
	routesManager.InitializeRoutes(router)
	log.Printf("Gin Server is running on http://%s:8080", host)
	err = router.Run(":8080")
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}

}
