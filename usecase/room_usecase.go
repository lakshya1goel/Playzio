package usecase

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/bootstrap/database"
	"github.com/lakshya1goel/Playzio/bootstrap/util"
	"github.com/lakshya1goel/Playzio/domain"
	"github.com/lakshya1goel/Playzio/domain/dto"
	"github.com/lakshya1goel/Playzio/domain/model"
	"github.com/lakshya1goel/Playzio/repository"
)

type RoomUsecase interface {
	CreateRoom(c *gin.Context, room model.Room) (*model.Room, *domain.HttpError)
	JoinRoom(c *gin.Context, joinCode string) (*model.Room, *domain.HttpError)
	GetAllPublicRooms(c *gin.Context) ([]model.Room, *domain.HttpError)
	LeaveRoom(c *gin.Context) *domain.HttpError
}

type roomUsecase struct {
	roomRepo       repository.RoomRepository
	userRepo       repository.UserRepository
	roomMemberRepo repository.RoomMemberRepository
}

func NewRoomUsecase() RoomUsecase {
	return &roomUsecase{
		roomRepo:       repository.NewRoomRepository(),
		userRepo:       repository.NewUserRepository(),
		roomMemberRepo: repository.NewRoomMemberRepository(),
	}
}

func (ru *roomUsecase) CreateRoom(c *gin.Context, room model.Room) (*model.Room, *domain.HttpError) {
	joinCode, err := util.GenerateRandomCode(6)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to generate join code",
		}
	}
	room.JoinCode = joinCode

	userInfo, userErr := ru.extractUserInfo(c)
	if userErr != nil {
		return nil, &domain.HttpError{
			StatusCode: userErr.StatusCode,
			Message:    userErr.Message,
		}
	}

	if userInfo.Type == "google" {
		room.CreatedBy = userInfo.UserID
	} else {
		room.CreatorGuestID = userInfo.GuestID
	}

	member := ru.createRoomMember(userInfo, room.ID, true)
	room.Members = []model.RoomMember{member}

	createdRoom, err := ru.roomRepo.CreateRoom(c, room)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to create room",
		}
	}

	err = database.Db.Preload("Members").First(&createdRoom, createdRoom.ID).Error
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to fetch created room with members",
		}
	}

	return &createdRoom, nil
}

func (ru *roomUsecase) JoinRoom(c *gin.Context, joinCode string) (*model.Room, *domain.HttpError) {
	room, err := ru.roomRepo.GetRoomByJoinCode(c, joinCode)
	if err != nil {
		if err.Error() == "record not found" {
			return nil, &domain.HttpError{
				StatusCode: http.StatusNotFound,
				Message:    "Room with the provided join code does not exist",
			}
		}
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to retrieve room",
		}
	}

	userInfo, userErr := ru.extractUserInfo(c)
	if userErr != nil {
		return nil, &domain.HttpError{
			StatusCode: userErr.StatusCode,
			Message:    userErr.Message,
		}
	}

	exists, existsErr := ru.isUserInRoom(c, userInfo, room.ID)
	if existsErr != nil {
		return nil, &domain.HttpError{
			StatusCode: existsErr.StatusCode,
			Message:    existsErr.Message,
		}
	}

	if exists {
		return nil, &domain.HttpError{
			StatusCode: http.StatusConflict,
			Message:    "User already joined the room",
		}
	}

	member := ru.createRoomMember(userInfo, room.ID, false)

	if err := ru.roomRepo.AddRoomMember(c, &member); err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to add member to room",
		}
	}

	room, err = ru.roomRepo.GetRoomByID(c, room.ID)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to fetch updated room with members",
		}
	}

	return &room, nil
}

func (ru *roomUsecase) GetAllPublicRooms(c *gin.Context) ([]model.Room, *domain.HttpError) {
	rooms, err := ru.roomRepo.GetAllPublicRooms(c)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to retrieve public rooms",
		}
	}

	if len(rooms) == 0 {
		return nil, &domain.HttpError{
			StatusCode: http.StatusNotFound,
			Message:    "No public rooms found",
		}
	}

	return rooms, nil
}

func (ru *roomUsecase) LeaveRoom(c *gin.Context) *domain.HttpError {
	userInfo, err := ru.extractUserInfo(c)
	if err != nil {
		return &domain.HttpError{
			StatusCode: err.StatusCode,
			Message:    err.Message,
		}
	}

	roomMember, err := ru.getRoomMemberByUserInfo(c, userInfo)
	if err != nil {
		return &domain.HttpError{
			StatusCode: err.StatusCode,
			Message:    err.Message,
		}
	}

	if roomMember.IsCreator {
		return ru.handleCreatorLeaving(c, userInfo, roomMember)
	} else {
		return ru.deleteRoomMember(c, userInfo, roomMember.RoomID)
	}
}

func (ru *roomUsecase) extractUserInfo(c *gin.Context) (*dto.User, *domain.HttpError) {
	userType := c.GetString("user_type")
	userInfo := &dto.User{Type: userType}

	switch userType {
	case "google":
		userID, exists := c.Get("user_id")
		if !exists {
			return nil, &domain.HttpError{
				StatusCode: http.StatusUnauthorized,
				Message:    "User not authenticated",
			}
		}
		uid := userID.(uint)
		userInfo.UserID = &uid

		username, exists := c.Get("user_name")
		if !exists {
			return nil, &domain.HttpError{
				StatusCode: http.StatusUnauthorized,
				Message:    "Username not found in context",
			}
		}
		usernameStr, ok := username.(string)
		if !ok {
			return nil, &domain.HttpError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Username type assertion failed",
			}
		}
		userInfo.Username = &usernameStr

	case "guest":
		guestID := c.GetString("guest_id")
		guestName := c.GetString("guest_name")
		userInfo.GuestID = &guestID
		userInfo.GuestName = &guestName

	default:
		return nil, &domain.HttpError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Invalid user type",
		}
	}

	return userInfo, nil
}

func (ru *roomUsecase) createRoomMember(userInfo *dto.User, roomID uint, isCreator bool) model.RoomMember {
	member := model.RoomMember{
		RoomID:    roomID,
		IsCreator: isCreator,
	}

	if userInfo.Type == "google" {
		member.UserID = userInfo.UserID
		member.Username = userInfo.Username
	} else {
		member.GuestID = userInfo.GuestID
		member.GuestName = userInfo.GuestName
	}

	return member
}

func (ru *roomUsecase) isUserInRoom(c *gin.Context, userInfo *dto.User, roomID uint) (bool, *domain.HttpError) {
	var exists bool
	var err error

	if userInfo.Type == "google" {
		exists, err = ru.roomRepo.IsUserInRoom(c, roomID, *userInfo.UserID)
		if err != nil {
			return false, &domain.HttpError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to check room membership",
			}
		}
	} else {
		exists, err = ru.roomRepo.IsGuestInRoom(c, roomID, *userInfo.GuestID)
		if err != nil {
			return false, &domain.HttpError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to check guest room membership",
			}
		}
	}

	return exists, nil
}

func (ru *roomUsecase) getRoomMemberByUserInfo(c *gin.Context, userInfo *dto.User) (model.RoomMember, *domain.HttpError) {
	var member model.RoomMember
	var err error

	if userInfo.Type == "google" {
		member, err = ru.roomMemberRepo.GetRoomMemberByUserID(c, *userInfo.UserID)
		if err != nil {
			return model.RoomMember{}, &domain.HttpError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to retrieve room for user",
			}
		}
	} else {
		member, err = ru.roomMemberRepo.GetRoomMemberByGuestID(c, *userInfo.GuestID)
		if err != nil {
			return model.RoomMember{}, &domain.HttpError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to retrieve room member for guest",
			}
		}
	}

	return member, nil
}

func (ru *roomUsecase) deleteRoomMember(c *gin.Context, userInfo *dto.User, roomID uint) *domain.HttpError {
	var err error

	if userInfo.Type == "google" {
		err = ru.roomMemberRepo.DeleteRoomMember(c, roomID, *userInfo.UserID)
	} else {
		err = ru.roomMemberRepo.DeleteRoomMemberByGuestID(c, roomID, *userInfo.GuestID)
	}

	if err != nil {
		return &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to delete room member",
		}
	}

	return nil
}

func (ru *roomUsecase) changeRoomCreator(c *gin.Context, roomID uint, newCreator model.RoomMember) *domain.HttpError {
	var err error

	if newCreator.UserID != nil {
		err = ru.roomRepo.ChangeRoomCreator(c, roomID, *newCreator.UserID)
	} else {
		err = ru.roomRepo.ChangeRoomGuestCreator(c, roomID, *newCreator.GuestID)
	}

	if err != nil {
		return &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to change room creator",
		}
	}

	return nil
}

func (ru *roomUsecase) handleCreatorLeaving(c *gin.Context, userInfo *dto.User, roomMember model.RoomMember) *domain.HttpError {
	members, err := ru.roomMemberRepo.GetRoomMembersByRoomID(c, roomMember.RoomID)
	if err != nil {
		return &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to retrieve members of the room",
		}
	}

	if len(members) == 1 {
		err := ru.roomRepo.DeleteRoom(c, roomMember.RoomID)
		if err != nil {
			return &domain.HttpError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to delete room",
			}
		}
	} else {
		newCreator := members[1]
		if httpErr := ru.changeRoomCreator(c, roomMember.RoomID, newCreator); httpErr != nil {
			return httpErr
		}

		err = ru.roomMemberRepo.UpdateRoomMemberToCreator(c, roomMember.RoomID, newCreator)
		if err != nil {
			return &domain.HttpError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to update room member to creator",
			}
		}
	}

	return ru.deleteRoomMember(c, userInfo, roomMember.RoomID)
}
