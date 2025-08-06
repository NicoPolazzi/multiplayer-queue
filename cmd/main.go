package main

import (
	"log"

	"github.com/NicoPolazzi/multiplayer-queue/internal/handlers"
	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	"github.com/NicoPolazzi/multiplayer-queue/internal/repository"
	"github.com/NicoPolazzi/multiplayer-queue/internal/routes"
	"github.com/NicoPolazzi/multiplayer-queue/internal/service"
	"github.com/NicoPolazzi/multiplayer-queue/internal/token"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {

	// TODO: Place the secret key in an environment variable
	jwtSecret := []byte("super-secret-key")

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("migration failed:", err)
	}

	userRepo := repository.NewSQLUserRepository(db)
	tokenManager := token.NewJWTTokenManager(jwtSecret)
	authService := service.NewJWTAuthService(userRepo)
	authService.(*service.JWTAuthService).SetTokenManager(tokenManager)
	userHandler := handlers.NewUserHandler(authService)
	routesManager := routes.NewRoutes(userHandler, &tokenManager)
	router := gin.Default()
	router.LoadHTMLGlob("web/templates/*")
	routesManager.InitializeRoutes(router)

	// Server is running on localhost port 8080 by default
	router.Run()
}
