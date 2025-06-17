package database

import (
	"fmt"
	"strconv"

	"github.com/lakshya1goel/Playzio/bootstrap"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Db *gorm.DB

func ConnectDb(env *bootstrap.Env) {
	host := env.DBHost
	portStr := env.DBPort
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic("Invalid port number: " + err.Error())
	}
	user := env.DBUser
	password := env.DBPass
	dbName := env.DBName
	configData := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)
	var dbErr error
	Db, dbErr = gorm.Open(postgres.Open(configData), &gorm.Config{})
	if dbErr != nil {
		panic("Error connecting to database: " + dbErr.Error())
	}

	fmt.Println("\x1b[32m...............Database connected..................\x1b[0m")

	InitDB()
}
