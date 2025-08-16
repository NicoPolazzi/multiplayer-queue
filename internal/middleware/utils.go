package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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
