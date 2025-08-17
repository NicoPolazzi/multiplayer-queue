package middleware

import (
	"github.com/NicoPolazzi/multiplayer-queue/internal/token"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	tokenManager token.TokenManager
}

func NewAuthMiddleware(tokenManager token.TokenManager) *AuthMiddleware {
	return &AuthMiddleware{tokenManager: tokenManager}
}

func (m *AuthMiddleware) CheckUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("is_logged_in", false)
		tokenString, err := ctx.Cookie("jwt")
		if err != nil {
			ctx.Next()
			return
		}

		username, err := m.tokenManager.Validate(tokenString)
		if err == token.ErrInvalidToken {
			ctx.Set("is_logged_in", false)
			ctx.Next()
			return
		}

		ctx.Set("is_logged_in", true)
		ctx.Set("username", username)
		ctx.Next()
	}
}
