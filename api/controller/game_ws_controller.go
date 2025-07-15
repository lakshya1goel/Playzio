package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/bootstrap/util"
	"github.com/lakshya1goel/Playzio/usecase"
	"github.com/lakshya1goel/Playzio/websocket"
)

type GameWSController struct {
	usecase usecase.GameWSUsecase
	pool    *websocket.GamePool
}

func NewGameWSController(pool *websocket.GamePool, wsUsecase usecase.GameWSUsecase) *GameWSController {
	return &GameWSController{
		usecase: wsUsecase,
		pool:    pool,
	}
}

func (wsc *GameWSController) HandleGameWebSocket(c *gin.Context) {
	userId, userName, conn, ok := util.UpgradeWithUserID(c)
	if !ok {
		return
	}

	client := &websocket.GameClient{
		BaseClient: websocket.BaseClient{
			Conn:     conn,
			UserId:   userId,
			UserName: userName,
		},
		Pool: wsc.pool,
	}

	go wsc.usecase.Read(client)
}
