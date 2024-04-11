package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"

	"server/db"
	"server/schemas"
)

func IsAuthorized(needAdmin bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("token")
		var user schemas.User

		if v, ok := db.UserCache.Get(token); ok {
			user, ok = v.(schemas.User)
			if !ok {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid user cache entry"})
			}
		} else {
			db.DB.Model(&schemas.User{}).First(&user, "token = ?", token)
			if user.ID == 0 {
				c.AbortWithStatus(http.StatusUnauthorized)
			}
		}

		db.UserCache.Set(token, user, cache.DefaultExpiration)
		if !user.IsAdmin && needAdmin {
			c.AbortWithStatus(http.StatusForbidden)
		}

		c.Next()
	}
}
