package middleware

import (
	"net/http"

	"github.com/NicoPolazzi/multiplayer-queue/internal/token"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(manager token.TokenManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString, err := ctx.Cookie("jwt")
		if err != nil {
			ctx.Redirect(http.StatusSeeOther, "/login")
			ctx.Abort()
			return
		}

		username, err := manager.Validate(tokenString)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error", "message": err.Error()})
			return
		}

		ctx.Set("username", username)
		ctx.Next()
	}
}
