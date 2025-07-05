package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/lakshya1goel/Playzio/api/controller"
	"github.com/lakshya1goel/Playzio/api/middleware"
)

func RoomRoutes(router *gin.RouterGroup, roomController *controller.RoomController) {
	roomRouter := router.Group("/room")
	roomRouter.Use(middleware.AuthMiddleware())
	{
		roomRouter.POST("/", roomController.CreateRoom)
		roomRouter.POST("/join", roomController.JoinRoom)
		roomRouter.GET("/public", roomController.GetAllPublicRooms)
		roomRouter.POST("/leave", roomController.LeaveRoom)
	}
}
