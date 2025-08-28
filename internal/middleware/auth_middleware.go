package middleware

import (
	"log"
	"net/http"

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
		tokenString, err := ctx.Cookie("token")
		if err != nil {
			ctx.Next()
			return
		}

		username, err := m.tokenManager.Validate(tokenString)
		if err != nil {
			if err != token.ErrInvalidToken {
				log.Printf("Error validating token: %v", err)
			}
			ctx.Next()
			return
		}

		setUserInContext(ctx, &User{Username: username})
		ctx.Next()
	}
}

func EnsureLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := UserFromContext(c); !ok {
			c.Redirect(http.StatusSeeOther, "/user/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

func EnsureNotLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := UserFromContext(c); ok {
			c.Redirect(http.StatusSeeOther, "/")
			c.Abort()
			return
		}
		c.Next()
	}
}
