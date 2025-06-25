package usecase

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/bootstrap/util"
	"github.com/lakshya1goel/Playzio/domain"
	"github.com/lakshya1goel/Playzio/domain/dto"
	"github.com/lakshya1goel/Playzio/domain/model"
	"github.com/lakshya1goel/Playzio/repository"
)

type RoomUsecase interface {
	CreateRoom(c *gin.Context, room model.Room) (*dto.CreateRoomResponse, *domain.HttpError)
	JoinRoom(c *gin.Context, joinCode string) *domain.HttpError
}

type roomUsecase struct {
	roomRepo repository.RoomRepository
	userRepo repository.UserRepository
}

func NewRoomUsecase() RoomUsecase {
	return &roomUsecase{
		roomRepo: repository.NewRoomRepository(),
		userRepo: repository.NewUserRepository(),
	}
}

func (ru *roomUsecase) CreateRoom(c *gin.Context, room model.Room) (*dto.CreateRoomResponse, *domain.HttpError) {
	joinCode, err := util.GenerateRandomCode(6)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to generate join code",
		}
	}
	room.JoinCode = joinCode

	userID, exists := c.Get("user_id")
	if !exists {
		return nil, &domain.HttpError{
			StatusCode: http.StatusUnauthorized,
			Message:    "User not authenticated",
		}
	}
	room.CreatedBy = userID.(uint)

	room.Members = []model.User{}
	user, err := ru.userRepo.GetUserByID(c, room.CreatedBy)
	if err != nil {
		if err.Error() == "record not found" {
			return nil, &domain.HttpError{
				StatusCode: http.StatusNotFound,
				Message:    "Creator not found",
			}
		}
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to retrieve creator information",
		}
	}
	room.Members = append(room.Members, user)

	resp, err := ru.roomRepo.CreateRoom(c, room)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to create room",
		}
	}

	user.RoomID = &room.ID
	err = ru.userRepo.UpdateUser(c, &user)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to update user",
		}
	}

	response := &dto.CreateRoomResponse{
		ID:        resp.ID,
		Name:      resp.Name,
		Type:      resp.Type,
		CreatedBy: resp.CreatedBy,
		JoinCode:  resp.JoinCode,
		Members:   resp.Members,
	}

	return response, nil
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

	userID, exists := c.Get("user_id")
	if !exists {
		return &domain.HttpError{
			StatusCode: http.StatusUnauthorized,
			Message:    "User not authenticated",
		}
	}

	user, err := ru.userRepo.GetUserByID(c, userID.(uint))
	if err != nil {
		if err.Error() == "record not found" {
			return &domain.HttpError{
				StatusCode: http.StatusNotFound,
				Message:    "User not found",
			}
		}
		return &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to retrieve user",
		}
	}

	room.Members = append(room.Members, user)
	err = ru.roomRepo.UpdateRoom(c, room)
	if err != nil {
		return &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to update room",
		}
	}

	user.RoomID = &room.ID
	err = ru.userRepo.UpdateUser(c, &user)
	if err != nil {
		return &domain.HttpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to update user",
		}
	}

	return nil
}
