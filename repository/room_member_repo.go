package repository

import (
	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/bootstrap/database"
	"github.com/lakshya1goel/Playzio/domain/model"
)

type RoomMemberRepository interface {
	GetRoomMemberByUserID(c *gin.Context, userID uint) (model.RoomMember, error)
	DeleteRoomMember(c *gin.Context, roomID uint, userID uint) error
	GetRoomMembersByRoomID(c *gin.Context, roomID uint) ([]model.RoomMember, error)
	UpdateRoomMemberToCreator(c *gin.Context, roomID uint, member model.RoomMember) error
	GetRoomMemberByGuestID(c *gin.Context, guestID string) (model.RoomMember, error)
	DeleteRoomMemberByGuestID(c *gin.Context, roomID uint, guestID string) error
}

type roomMemberRepository struct{}

func NewRoomMemberRepository() RoomMemberRepository {
	return &roomMemberRepository{}
}

func (r *roomMemberRepository) GetRoomMemberByUserID(c *gin.Context, userID uint) (model.RoomMember, error) {
	var member model.RoomMember
	if err := database.Db.Where("user_id = ?", userID).Preload("User").First(&member).Error; err != nil {
		return model.RoomMember{}, err
	}
	return member, nil
}

func (r *roomMemberRepository) DeleteRoomMember(c *gin.Context, roomID uint, userID uint) error {
	if err := database.Db.Where("room_id = ? AND user_id = ?", roomID, userID).Delete(&model.RoomMember{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *roomMemberRepository) GetRoomMembersByRoomID(c *gin.Context, roomID uint) ([]model.RoomMember, error) {
	var members []model.RoomMember
	if err := database.Db.Where("room_id = ?", roomID).Preload("User").Find(&members).Error; err != nil {
		return []model.RoomMember{}, err
	}
	return members, nil
}

func (r *roomMemberRepository) UpdateRoomMemberToCreator(c *gin.Context, roomID uint, member model.RoomMember) error {
	if member.UserID != nil {
		if err := database.Db.Model(&model.RoomMember{}).
			Where("room_id = ? AND user_id = ?", roomID, member.UserID).
			Update("is_creator", true).Error; err != nil {
			return err
		}
	} else {
		if err := database.Db.Model(&model.RoomMember{}).
			Where("room_id = ? AND guest_id = ?", roomID, member.GuestID).
			Update("is_creator", true).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *roomMemberRepository) GetRoomMemberByGuestID(c *gin.Context, guestID string) (model.RoomMember, error) {
	var member model.RoomMember
	if err := database.Db.Where("guest_id = ?", guestID).Preload("User").First(&member).Error; err != nil {
		return model.RoomMember{}, err
	}
	return member, nil
}

func (r *roomMemberRepository) DeleteRoomMemberByGuestID(c *gin.Context, roomID uint, guestID string) error {
	if err := database.Db.Where("room_id = ? AND guest_id = ?", roomID, guestID).Delete(&model.RoomMember{}).Error; err != nil {
		return err
	}
	return nil
}
