package repository

import (
	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/bootstrap/database"
	"github.com/lakshya1goel/Playzio/domain/model"
)

type RoomRepository interface {
	CreateRoom(c *gin.Context, room model.Room) (model.Room, error)
	GetRoomByID(c *gin.Context, id uint) (model.Room, error)
}

type roomRepository struct{}

func NewRoomRepository() RoomRepository {
	return &roomRepository{}
}

func (r *roomRepository) CreateRoom(c *gin.Context, room model.Room) (model.Room, error) {
	if err := database.Db.Create(&room).Error; err != nil {
		return model.Room{}, err
	}
	return room, nil
}

func (r *roomRepository) GetRoomByID(c *gin.Context, id uint) (model.Room, error) {
	var room model.Room
	if err := database.Db.First(&room, id).Error; err != nil {
		return model.Room{}, err
	}
	return room, nil
}
