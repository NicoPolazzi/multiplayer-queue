package middleware

import (
	"net/http"
	"strings"

	"github.com/NicoPolazzi/multiplayer-queue/internal/token"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(manager token.TokenManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error",
				"message": "Invalid authorization format. Use: Bearer {token}"})
			return
		}
		tokenString := strings.TrimPrefix(auth, "Bearer ")
		username, err := manager.Validate(tokenString)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error", "message": err.Error()})
			return
		}

		ctx.Set("username", username)
		ctx.Next()
	}
}
