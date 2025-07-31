package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/bootstrap/util"
	"github.com/lakshya1goel/Playzio/websocket"
)

type ChatWSController struct {
	handler websocket.ChatHandler
	pool    *websocket.ChatPool
}

func NewChatWSController(pool *websocket.ChatPool, handler websocket.ChatHandler) *ChatWSController {
	return &ChatWSController{
		handler: handler,
		pool:    pool,
	}
}

func (wsc *ChatWSController) HandleWebSocket(c *gin.Context) {
	userId, _, conn, ok := util.UpgradeWithUserID(c)
	if !ok {
		return
	}

	client := &websocket.ChatClient{
		BaseClient: websocket.BaseClient{
			Conn:   conn,
			UserId: userId,
		},
		Pool: wsc.pool,
	}

	go wsc.handler.Read(client)
}
