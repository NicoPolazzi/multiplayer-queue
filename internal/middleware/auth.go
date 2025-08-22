package middleware

import (
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
		ctx.Set("is_logged_in", false)
		tokenString, err := ctx.Cookie("token")
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

func EnsureLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		if is, exists := c.Get("is_logged_in"); !exists || !is.(bool) {
			c.Redirect(http.StatusSeeOther, "/user/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

func EnsureNotLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		if is, exists := c.Get("is_logged_in"); exists && is.(bool) {
			c.Redirect(http.StatusSeeOther, "/")
			c.Abort()
			return
		}
		c.Next()
	}
}
