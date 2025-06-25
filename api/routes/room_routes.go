package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/lakshya1goel/Playzio/api/controller"
)

func RoomRoutes(router *gin.RouterGroup, roomController *controller.RoomController) {
	roomRouter := router.Group("/room")
	{
		roomRouter.POST("/", roomController.CreateRoom)
	}
}
