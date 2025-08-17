package main

import (
	"log"
	"os"

	"github.com/NicoPolazzi/multiplayer-queue/internal/handlers"
	"github.com/NicoPolazzi/multiplayer-queue/internal/middleware"
	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	repository "github.com/NicoPolazzi/multiplayer-queue/internal/repository/user"
	"github.com/NicoPolazzi/multiplayer-queue/internal/routes"
	"github.com/NicoPolazzi/multiplayer-queue/internal/service"
	"github.com/NicoPolazzi/multiplayer-queue/internal/token"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Check values in .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
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

	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("migration failed:", err)
	}

	userRepo := repository.NewSQLUserRepository(db)
	tokenManager := token.NewJWTTokenManager(key)
	authService := service.NewJWTAuthService(userRepo, tokenManager)
	userHandler := handlers.NewUserHandler(authService)
	authMiddleware := middleware.NewAuthMiddleware(tokenManager)
	routesManager := routes.NewRoutes(userHandler, authMiddleware)

	// Server is running on localhost port 8080 by default
	router := gin.Default()
	router.LoadHTMLGlob("web/templates/*")
	routesManager.InitializeRoutes(router)

	err = router.Run()
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
	log.Println("Server is running on http://localhost:8080")
}
