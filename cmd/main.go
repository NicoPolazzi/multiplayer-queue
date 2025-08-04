package main

import (
	"log"
	"net/http"
	"os"

	"github.com/NicoPolazzi/multiplayer-queue/internal/handler"
	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	"github.com/NicoPolazzi/multiplayer-queue/internal/repository"
	"github.com/NicoPolazzi/multiplayer-queue/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// TODO: change it with .env file
	jwtSecret := []byte("super-secret-key")

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("migration failed:", err)
	}

	userRepo := repository.NewSQLUserRepository(db)
	authService := service.NewJWTAuthService(userRepo, jwtSecret)
	authHandler := handler.NewAuthHandler(authService)

	r := gin.Default()
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server listening on port %s", port)
	if err := r.Run(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatal("server error:", err)
	}
}
