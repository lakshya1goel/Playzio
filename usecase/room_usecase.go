package usecase

import (
	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/bootstrap/util"
	"github.com/lakshya1goel/Playzio/domain"
	"github.com/lakshya1goel/Playzio/domain/dto"
	"github.com/lakshya1goel/Playzio/domain/model"
	"github.com/lakshya1goel/Playzio/repository"
)

type RoomUsecase interface {
	CreateRoom(c *gin.Context, room model.Room) (*dto.CreateRoomResponse, *domain.HttpError)
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
			StatusCode: 500,
			Message:    "Failed to generate join code",
		}
	}
	room.JoinCode = joinCode

	userID, exists := c.Get("user_id")
	if !exists {
		return nil, &domain.HttpError{
			StatusCode: 401,
			Message:    "User not authenticated",
		}
	}
	room.CreatedBy = userID.(uint)

	room.Members = []model.User{}
	user, err := ru.userRepo.GetUserByID(c, room.CreatedBy)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: 500,
			Message:    "Failed to retrieve creator information",
		}
	}
	room.Members = append(room.Members, user)

	resp, err := ru.roomRepo.CreateRoom(c, room)
	if err != nil {
		return nil, &domain.HttpError{
			StatusCode: 500,
			Message:    "Failed to create room",
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
