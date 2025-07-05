package repository

import (
	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/bootstrap/database"
	"github.com/lakshya1goel/Playzio/domain/model"
)

type RoomRepository interface {
	CreateRoom(c *gin.Context, room model.Room) (model.Room, error)
	GetRoomByID(c *gin.Context, id uint) (model.Room, error)
	GetRoomByJoinCode(c *gin.Context, joinCode string) (model.Room, error)
	UpdateRoom(c *gin.Context, room model.Room) error
	AddRoomMember(c *gin.Context, member *model.RoomMember) error
	IsUserInRoom(c *gin.Context, roomID uint, userID uint) (bool, error)
	IsGuestInRoom(c *gin.Context, roomID uint, guestID string) (bool, error)
	GetAllPublicRooms(c *gin.Context) ([]model.Room, error)
	DeleteRoom(c *gin.Context, roomID uint) error
	ChangeRoomCreator(c *gin.Context, roomID uint, creatorID uint) error
	ChangeRoomGuestCreator(c *gin.Context, roomID uint, guestID string) error
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

func (r *roomRepository) GetRoomByJoinCode(c *gin.Context, joinCode string) (model.Room, error) {
	var room model.Room
	if err := database.Db.Where("join_code = ?", joinCode).First(&room).Error; err != nil {
		return model.Room{}, err
	}
	return room, nil
}

func (r *roomRepository) UpdateRoom(c *gin.Context, room model.Room) error {
	if err := database.Db.Save(&room).Error; err != nil {
		return err
	}
	return nil
}

func (r *roomRepository) AddRoomMember(c *gin.Context, member *model.RoomMember) error {
	if err := database.Db.Create(member).Error; err != nil {
		return err
	}
	return nil
}

func (r *roomRepository) IsUserInRoom(c *gin.Context, roomID uint, userID uint) (bool, error) {
	var count int64
	err := database.Db.Model(&model.RoomMember{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *roomRepository) IsGuestInRoom(c *gin.Context, roomID uint, guestID string) (bool, error) {
	var count int64
	err := database.Db.Model(&model.RoomMember{}).
		Where("room_id = ? AND guest_id = ?", roomID, guestID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *roomRepository) GetAllPublicRooms(c *gin.Context) ([]model.Room, error) {
	var rooms []model.Room
	if err := database.Db.Preload("Members").Where("type = ?", "public").Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *roomRepository) DeleteRoom(c *gin.Context, roomID uint) error {
	if err := database.Db.Where("id = ?", roomID).Delete(&model.Room{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *roomRepository) ChangeRoomCreator(c *gin.Context, roomID uint, creatorID uint) error {
	if err := database.Db.Model(&model.Room{}).
		Where("id = ?", roomID).
		Update("creator_id", creatorID).Error; err != nil {
		return err
	}
	return nil
}

func (r *roomRepository) ChangeRoomGuestCreator(c *gin.Context, roomID uint, guestID string) error {
	if err := database.Db.Model(&model.Room{}).
		Where("id = ?", roomID).
		Update("creator_guest_id", guestID).Error; err != nil {
		return err
	}
	return nil
}
