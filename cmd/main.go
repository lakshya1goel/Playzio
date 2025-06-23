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
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	fmt.Println("hi " + os.Getenv("GOOGLE_CLIENT_ID"))
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

	router := gin.Default()
	apiRouter := router.Group("/api")
	{
		routes.AuthRoutes(apiRouter, controller.NewAuthController())
	}

	router.Run(":8000")
}
