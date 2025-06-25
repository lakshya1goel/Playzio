package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/lakshya1goel/Playzio/api/controller"
)

func WsRoutes(router *gin.RouterGroup, chatWsController *controller.ChatWSController, gameWsController *controller.GameWSController) {
	wsRouter := router.Group("/ws")
	{
		wsRouter.GET("/chat", chatWsController.HandleWebSocket)
		wsRouter.GET("/game", gameWsController.HandleGameWebSocket)
	}
}
