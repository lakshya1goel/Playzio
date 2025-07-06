package util

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func UpgradeWithUserID(c *gin.Context) (uint, string, *websocket.Conn, bool) {
	userType, exists := c.Get("user_type")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return 0, "", nil, false
	}

	userNameRaw, exists := c.Get("user_name")
	userName := ""
	if exists {
		if name, ok := userNameRaw.(string); ok {
			userName = name
		}
	}

	var userId uint

	switch userType.(string) {
	case "google":
		userIdRaw, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Missing user ID"})
			return 0, "", nil, false
		}
		if id, ok := userIdRaw.(uint); ok {
			userId = id
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid user ID"})
			return 0, "", nil, false
		}

	case "guest":
		guestIDRaw, exists := c.Get("guest_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Missing guest ID"})
			return 0, "", nil, false
		}
		if guestID, ok := guestIDRaw.(string); ok {
			userId = 0
			userName = userName + "_" + guestID
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid guest ID"})
			return 0, "", nil, false
		}

	default:
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid user type"})
		return 0, "", nil, false
	}

	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return 0, "", nil, false
	}

	return userId, userName, conn, true
}
