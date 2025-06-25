package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/lakshya1goel/Playzio/domain/model"
	"github.com/lakshya1goel/Playzio/usecase"
)

type WSController struct {
	usecase usecase.WSUsecase
	pool    *model.Pool
}

func NewWSController(pool *model.Pool, wsUsecase usecase.WSUsecase) *WSController {
	return &WSController{
		usecase: wsUsecase,
		pool:    pool,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (wsc *WSController) HandleWebSocket(c *gin.Context) {
	userIdFloat := float64(1)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &model.Client{
		Conn:   conn,
		Pool:   wsc.pool,
		UserId: uint(userIdFloat),
	}

	go wsc.usecase.Read(client)
}
