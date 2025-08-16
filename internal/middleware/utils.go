package middleware

import (
	"net/http"

	"github.com/NicoPolazzi/multiplayer-queue/internal/handlers"
	"github.com/gin-gonic/gin"
)

func EnsureLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		if is, exists := c.Get("is_logged_in"); !exists || !is.(bool) {
			c.Redirect(http.StatusSeeOther, handlers.LoginPagePath)
			c.Abort()
			return
		}
		c.Next()
	}
}

func EnsureNotLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		if is, exists := c.Get("is_logged_in"); exists && is.(bool) {
			c.Redirect(http.StatusSeeOther, handlers.IndexPagePath)
			c.Abort()
			return
		}
		c.Next()
	}
}
