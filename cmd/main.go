package main

import (
	"log"
	"net/http"
	"os"

	"github.com/NicoPolazzi/multiplayer-queue/internal/handler"
	"github.com/NicoPolazzi/multiplayer-queue/internal/middleware"
	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	"github.com/NicoPolazzi/multiplayer-queue/internal/repository"
	"github.com/NicoPolazzi/multiplayer-queue/internal/service"
	"github.com/NicoPolazzi/multiplayer-queue/internal/token"
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
	tokenManager := token.NewJWTTokenManager(jwtSecret)
	authService := service.NewJWTAuthService(userRepo)
	authService.(*service.JWTAuthService).SetTokenManager(tokenManager)
	authHandler := handler.NewAuthHandler(authService)

	r := gin.Default()
	r.LoadHTMLGlob("web/templates/*.html")

	r.GET("/", func(c *gin.Context) {
		if _, err := c.Cookie("jwt"); err == nil {
			c.Redirect(http.StatusSeeOther, "/dashboard")
			return
		}

		c.HTML(http.StatusOK, "base.html", gin.H{
			"title":    "Home",
			"template": "home",
		})
	})

	r.GET("/login", authHandler.ShowLogin)
	r.GET("/register", authHandler.ShowRegister)
	r.POST("/login", authHandler.Login)
	r.POST("/register", authHandler.Register)

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware(tokenManager))
	{
		protected.GET("/dashboard", func(c *gin.Context) {
			username := c.GetString("username")
			c.HTML(http.StatusOK, "base.html", gin.H{
				"title":    "Dashboard - Multiplayer Queue",
				"username": username,
				"template": "dashboard",
			})
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server listening on port %s", port)
	if err := r.Run(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatal("server error:", err)
	}
}
