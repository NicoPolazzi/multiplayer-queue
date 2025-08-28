package middleware

import "github.com/gin-gonic/gin"

type contextKey string

const userKey contextKey = "user"

type User struct {
	Username string
}

func setUserInContext(c *gin.Context, user *User) {
	c.Set(string(userKey), user)
}

func UserFromContext(c *gin.Context) (*User, bool) {
	user, exists := c.Get(string(userKey))
	if !exists {
		return nil, false
	}
	u, ok := user.(*User)
	return u, ok
}
