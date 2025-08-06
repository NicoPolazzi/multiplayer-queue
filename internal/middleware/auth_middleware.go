package middleware

import (
	"net/http"

	"github.com/NicoPolazzi/multiplayer-queue/internal/token"
	"github.com/gin-gonic/gin"
)

func CheckUser(manager token.TokenManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("is_logged_in", false)

		tokenString, err := ctx.Cookie("jwt")
		if err != nil {
			ctx.Next()
			return
		}

		username, err := manager.Validate(tokenString)
		if err != nil {
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
