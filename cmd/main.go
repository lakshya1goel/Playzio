package main

import (
	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/bootstrap"
	"github.com/lakshya1goel/Playzio/bootstrap/database"
)

func main() {
	app := bootstrap.App()
	env := app.Env

	database.ConnectDb(env)

	router := gin.Default()
	router.Run(":8000")
}
