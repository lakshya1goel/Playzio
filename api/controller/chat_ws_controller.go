package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/lakshya1goel/Playzio/domain/model"
	"github.com/lakshya1goel/Playzio/usecase"
)

type ChatWSController struct {
	usecase usecase.ChatWSUsecase
	pool    *model.ChatPool
}

func NewChatWSController(pool *model.ChatPool, wsUsecase usecase.ChatWSUsecase) *ChatWSController {
	return &ChatWSController{
		usecase: wsUsecase,
		pool:    pool,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (wsc *ChatWSController) HandleWebSocket(c *gin.Context) {
	userIdRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	userIdFloat, ok := userIdRaw.(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid user ID in context"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &model.ChatClient{
		Conn:   conn,
		Pool:   wsc.pool,
		UserId: uint(userIdFloat),
	}

	go wsc.usecase.Read(client)
}
