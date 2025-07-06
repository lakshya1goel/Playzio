package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/bootstrap/util"
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

func (wsc *ChatWSController) HandleWebSocket(c *gin.Context) {
	userId, _, conn, ok := util.UpgradeWithUserID(c)
	if !ok {
		return
	}

	client := &model.ChatClient{
		BaseClient: model.BaseClient{
			Conn:   conn,
			UserId: userId,
		},
		Pool: wsc.pool,
	}

	go wsc.usecase.Read(client)
}
