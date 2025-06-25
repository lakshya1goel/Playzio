package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/lakshya1goel/Playzio/api/controller"
)

func GameWsRoutes(router *gin.RouterGroup, gameWsController *controller.GameWSController) {
	gameWsRouter := router.Group("/ws/game")
	{
		gameWsRouter.GET("/", gameWsController.HandleGameWebSocket)
	}
}
