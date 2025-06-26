package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/bootstrap/util"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Missing authentication token"})
			c.Abort()
			return
		}

		tokenParts := strings.Split(tokenString, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authentication token"})
			c.Abort()
			return
		}

		tokenString = tokenParts[1]

		claims, err := util.VerifyToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authentication token"})
			c.Abort()
			return
		}

		if claims["type"] == "guest" {
			c.Set("user_type", "guest")
			c.Set("guest_name", claims["name"])
			c.Set("guest_id", claims["guest_id"].(string))
		} else {
			if userID, ok := claims["user_id"].(float64); ok {
				c.Set("user_type", "google")
				c.Set("user_id", uint(userID))
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token (missing user_id)"})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
