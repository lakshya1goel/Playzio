package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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

	client := &model.GameClient{
		Conn:   conn,
		Pool:   wsc.pool,
		UserId: uint(userIdFloat),
	}

	go wsc.usecase.Read(client)
}
