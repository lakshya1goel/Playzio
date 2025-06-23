package main

import (
	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/bootstrap"
	"github.com/lakshya1goel/Playzio/bootstrap/database"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)

func main() {
	goth.UseProviders(
		google.New(
			"GOOGLE_CLIENT_ID",
			"GOOGLE_CLIENT_SECRET",
			"http://localhost:8000/auth/google/callback",
		),
	)

	app := bootstrap.App()
	env := app.Env

	database.ConnectDb(env)

	router := gin.Default()
	router.Run(":8000")
}
