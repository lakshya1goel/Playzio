package database

import (
	"fmt"

	"github.com/lakshya1goel/Playzio/domain/model"
	_ "github.com/lib/pq"
)

func InitDB() error {
	if Db == nil {
		return fmt.Errorf("database connection not established. Call ConnectDb first")
	}

	err := Db.AutoMigrate(&model.User{})
	if err != nil {
		return fmt.Errorf("error creating expenses table: %v", err)
	}

	fmt.Println("Database tables initialized successfully")
	return nil
}
