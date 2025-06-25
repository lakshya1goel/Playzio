package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/lakshya1goel/Playzio/api/controller"
)

func WsRoutes(router *gin.RouterGroup, wsController *controller.WSController) {
	wsRouter := router.Group("/ws")
	{
		wsRouter.GET("/", wsController.HandleWebSocket)
	}
}
