package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/bootstrap/util"
	"github.com/lakshya1goel/Playzio/domain/model"
	"github.com/lakshya1goel/Playzio/usecase"
)

type GameWSController struct {
	usecase usecase.GameWSUsecase
	pool    *model.GamePool
}

func NewGameWSController(pool *model.GamePool, wsUsecase usecase.GameWSUsecase) *GameWSController {
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

	client := &model.GameClient{
		BaseClient: model.BaseClient{
			Conn:     conn,
			UserId:   userId,
			UserName: userName,
		},
		Pool: wsc.pool,
	}

	go wsc.usecase.Read(client)
}
