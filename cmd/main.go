package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/lakshya1goel/Playzio/api/controller"
	"github.com/lakshya1goel/Playzio/api/routes"
	"github.com/lakshya1goel/Playzio/bootstrap"
	"github.com/lakshya1goel/Playzio/bootstrap/database"
	"github.com/lakshya1goel/Playzio/bootstrap/util"
	"github.com/lakshya1goel/Playzio/usecase"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	goth.UseProviders(
		google.New(
			os.Getenv("GOOGLE_CLIENT_ID"),
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			os.Getenv("GOOGLE_REDIRECT_URI"),
			"email", "profile",
		),
	)

	app := bootstrap.App()
	env := app.Env

	database.ConnectDb(env)
	util.InitGoogleOAuth()

	router := gin.Default()

	authController := controller.NewAuthController()
	gameController := controller.NewGameWSController(app.GamePool)
	chatController := controller.NewChatWSController(app.ChatPool, usecase.NewChatWSUsecase())
	roomController := controller.NewRoomController()

        router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to Playzio",
		})
	})
	apiRouter := router.Group("/api")
	{
		routes.AuthRoutes(apiRouter, authController)
		routes.WsRoutes(apiRouter, chatController, gameController)
		routes.RoomRoutes(apiRouter, roomController)
	}

	router.Run(":8000")
}
