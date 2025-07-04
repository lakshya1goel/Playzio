package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/lakshya1goel/Playzio/api/controller"
)

func AuthRoutes(router *gin.RouterGroup, authController *controller.AuthController) {
	authRouter := router.Group("/auth")
	{
		authRouter.GET("/", authController.GoogleSignIn)
		authRouter.POST("/callback", authController.GoogleCallback)
		authRouter.GET("/google", authController.BeginAuth)
		// authRouter.GET("/callback", authController.Callback)
		authRouter.POST("/guest", authController.GuestAuth)
		authRouter.POST("/access-token", authController.GetAccessTokenFromRefreshToekn)
	}
}
