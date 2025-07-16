package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/bootstrap/util"
	"github.com/lakshya1goel/Playzio/usecase"
	"github.com/lakshya1goel/Playzio/websocket"
)

type ChatWSController struct {
	usecase usecase.ChatWSUsecase
	pool    *websocket.ChatPool
}

func NewChatWSController(pool *websocket.ChatPool, wsUsecase usecase.ChatWSUsecase) *ChatWSController {
	return &ChatWSController{
		usecase: wsUsecase,
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

	go wsc.usecase.Read(client)
}
