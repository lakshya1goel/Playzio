package usecase

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/bootstrap/database"
	"github.com/lakshya1goel/Playzio/bootstrap/util"
	"github.com/lakshya1goel/Playzio/domain"
	"github.com/lakshya1goel/Playzio/domain/model"
	"github.com/lakshya1goel/Playzio/repository"
)

type RoomUsecase interface {
	CreateRoom(c *gin.Context, room model.Room) (*model.Room, *domain.HttpError)
	JoinRoom(c *gin.Context, joinCode string) *domain.HttpError
	GetAllPublicRooms(c *gin.Context) ([]model.Room, *domain.HttpError)
	LeaveRoom(c *gin.Context) *domain.HttpError
}

type roomUsecase struct {
	roomRepo repository.RoomRepository
	userRepo repository.UserRepository
	roomMemberRepo repository.RoomMemberRepository
}

func NewRoomUsecase() RoomUsecase {
	return &roomUsecase{
		roomRepo: repository.NewRoomRepository(),
		userRepo: repository.NewUserRepository(),
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

	userType := c.GetString("user_type")
	var members []model.RoomMember

	if userType == "google" {
		userID, exists := c.Get("user_id")
		if !exists {
			return nil, &domain.HttpError{
				StatusCode: http.StatusUnauthorized,
				Message:    "User not authenticated",
			}
		}
		uid := userID.(uint)
		room.CreatedBy = &uid

		members = append(members, model.RoomMember{
			UserID:    &uid,
			IsCreator: true,
		})
	} else if userType == "guest" {
		guestID := c.GetString("guest_id")
		guestName := c.GetString("guest_name")
		room.CreatorGuestID = &guestID

		members = append(members, model.RoomMember{
			GuestID:   &guestID,
			GuestName: &guestName,
			IsCreator: true,
		})
	} else {
		return nil, &domain.HttpError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Invalid user type",
		}
	}

	room.Members = members
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

func (ru *roomUsecase) JoinRoom(c *gin.Context, joinCode string) *domain.HttpError {
	room, err := ru.roomRepo.GetRoomByJoinCode(c, joinCode)
	if err != nil {
		if err.Error() == "record not found" {
			return &domain.HttpError{
				StatusCode: http.StatusNotFound,
				Message:    "Room with the provided join code does not exist",
			}
		}
		return &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to retrieve room",
		}
	}

	userType := c.GetString("user_type")
	var member model.RoomMember

	if userType == "google" {
		userID, exists := c.Get("user_id")
		if !exists {
			return &domain.HttpError{
				StatusCode: http.StatusUnauthorized,
				Message:    "User not authenticated",
			}
		}
		uid := userID.(uint)

		exists, err := ru.roomRepo.IsUserInRoom(c, room.ID, uid)
		if err != nil {
			return &domain.HttpError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to check room membership",
			}
		}
		if exists {
			return &domain.HttpError{
				StatusCode: http.StatusConflict,
				Message:    "User already joined the room",
			}
		}

		member = model.RoomMember{
			RoomID: room.ID,
			UserID: &uid,
		}
	} else if userType == "guest" {
		guestID := c.GetString("guest_id")
		guestName := c.GetString("guest_name")

		exists, err := ru.roomRepo.IsGuestInRoom(c, room.ID, guestID)
		if err != nil {
			return &domain.HttpError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to check guest room membership",
			}
		}
		if exists {
			return &domain.HttpError{
				StatusCode: http.StatusConflict,
				Message:    "Guest already joined the room",
			}
		}

		member = model.RoomMember{
			RoomID:    room.ID,
			GuestID:   &guestID,
			GuestName: &guestName,
		}
	} else {
		return &domain.HttpError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Invalid user type",
		}
	}

	if err := ru.roomRepo.AddRoomMember(c, &member); err != nil {
		return &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to add member to room",
		}
	}

	return nil
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
	userType := c.GetString("user_type")

	if userType == "google" {
		userID, exists := c.Get("user_id")
		if !exists {
			return &domain.HttpError{
				StatusCode: http.StatusUnauthorized,
				Message:    "User not authenticated",
			}
		}
		uid := userID.(uint)

		roomMember, err := ru.roomMemberRepo.GetRoomMemberByUserID(c, uid)
		if err != nil {
			return &domain.HttpError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to retrieve room for creator user",
			}
		}

		if roomMember.IsCreator {
			resp, err := ru.roomMemberRepo.GetRoomMembersByRoomID(c, roomMember.RoomID)
			if err != nil {
				return &domain.HttpError{
					StatusCode: http.StatusInternalServerError,
					Message:    "Failed to retrieve members of the room",
				}
			}
			if len(resp) == 1 {
				err := ru.roomRepo.DeleteRoom(c, roomMember.RoomID)
				if err != nil {
					return &domain.HttpError{
						StatusCode: http.StatusInternalServerError,
						Message:    "Failed to delete room",
					}
				}
			} else {
				err = ru.roomRepo.ChangeRoomCreator(c, roomMember.RoomID, *resp[1].UserID)
				if err != nil {
					return &domain.HttpError{
						StatusCode: http.StatusInternalServerError,
						Message:    "Failed to change room creator",
					}
				}
				err = ru.roomMemberRepo.UpdateRoomMemberToCreator(c, roomMember.RoomID, resp[1])
				if err != nil {
					return &domain.HttpError{
						StatusCode: http.StatusInternalServerError,
						Message:    "Failed to update room member to creator",
					}
				}
			}
			err = ru.roomMemberRepo.DeleteRoomMember(c, roomMember.RoomID, uid)
			if err != nil {
				return &domain.HttpError{
					StatusCode: http.StatusInternalServerError,
					Message:    "Failed to delete room member",
				}
			}
		} else {
			err := ru.roomMemberRepo.DeleteRoomMember(c, roomMember.RoomID, uid)
			if err != nil {
				return &domain.HttpError{
					StatusCode: http.StatusInternalServerError,
					Message:    "Failed to delete room member",
				}
			}
		}
	} else if userType == "guest" {
		guestID := c.GetString("guest_id")
		roomMember, err := ru.roomMemberRepo.GetRoomMemberByGuestID(c, guestID)
		if err != nil {
			return &domain.HttpError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to retrieve room member for guest",
			}
		}

		if roomMember.IsCreator {
			resp, err := ru.roomMemberRepo.GetRoomMembersByRoomID(c, roomMember.RoomID)
			if err != nil {
				return &domain.HttpError{
					StatusCode: http.StatusInternalServerError,
					Message:    "Failed to retrieve members of the room",
				}
			}
			if len(resp) == 1 {
				err := ru.roomRepo.DeleteRoom(c, roomMember.RoomID)
				if err != nil {
					return &domain.HttpError{
						StatusCode: http.StatusInternalServerError,
						Message:    "Failed to delete room",
					}
				}
			} else {
				err = ru.roomRepo.ChangeRoomGuestCreator(c, roomMember.RoomID, *resp[1].GuestID)
				if err != nil {
					return &domain.HttpError{
						StatusCode: http.StatusInternalServerError,
						Message:    "Failed to change room creator",
					}
				}
				err = ru.roomMemberRepo.UpdateRoomMemberToCreator(c, roomMember.RoomID, resp[1])
				if err != nil {
					return &domain.HttpError{
						StatusCode: http.StatusInternalServerError,
						Message:    "Failed to update room member to creator",
					}
				}
			}
			err = ru.roomMemberRepo.DeleteRoomMemberByGuestID(c, roomMember.RoomID, guestID)
			if err != nil {
				return &domain.HttpError{
					StatusCode: http.StatusInternalServerError,
					Message:    "Failed to delete room member",
				}
			}
		} else {
			err := ru.roomMemberRepo.DeleteRoomMemberByGuestID(c, roomMember.RoomID, guestID)
			if err != nil {
				return &domain.HttpError{
					StatusCode: http.StatusInternalServerError,
					Message:    "Failed to delete room member",
				}
			}
		}
	} else {
		return &domain.HttpError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Invalid user type",
		}
	}
	return nil
}
