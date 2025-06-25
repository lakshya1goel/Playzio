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

func UpgradeWithUserID(c *gin.Context) (uint, *websocket.Conn, bool) {
	userIdRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return 0, nil, false
	}

	userIdFloat, ok := userIdRaw.(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid user ID in context"})
		return 0, nil, false
	}

	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return 0, nil, false
	}

	return uint(userIdFloat), conn, true
}
