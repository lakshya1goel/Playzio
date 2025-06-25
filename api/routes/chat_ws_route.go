package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/lakshya1goel/Playzio/api/controller"
)

func ChatWsRoutes(router *gin.RouterGroup, chatWsController *controller.ChatWSController) {
	chatWsRouter := router.Group("/ws/chat")
	{
		chatWsRouter.GET("/", chatWsController.HandleWebSocket)
	}
}
