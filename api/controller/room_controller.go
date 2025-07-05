package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lakshya1goel/Playzio/domain"
	"github.com/lakshya1goel/Playzio/domain/dto"
	"github.com/lakshya1goel/Playzio/domain/model"
	"github.com/lakshya1goel/Playzio/usecase"
)

type RoomController struct {
	roomUsecase usecase.RoomUsecase
}

func NewRoomController() *RoomController {
	return &RoomController{
		roomUsecase: usecase.NewRoomUsecase(),
	}
}

func (rc *RoomController) CreateRoom(c *gin.Context) {
	var request dto.CreateRoomRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Message: "Invalid request data",
		})
		return
	}

	room := model.Room{
		Name: request.Name,
		Type: request.Type,
	}

	response, err := rc.roomUsecase.CreateRoom(c, room)

	if err != nil {
		c.JSON(err.StatusCode, domain.ErrorResponse{
			Message: err.Message,
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Message: "Room created successfully!",
		Data:    response,
	})
}

func (rc *RoomController) JoinRoom(c *gin.Context) {
	joinCode := c.Query("join_code")
	if joinCode == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Message: "Join code is required",
		})
		return
	}

	err := rc.roomUsecase.JoinRoom(c, joinCode)

	if err != nil {
		c.JSON(err.StatusCode, domain.ErrorResponse{
			Message: err.Message,
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Message: "Room joined successfully!",
	})
}

func (rc *RoomController) GetAllPublicRooms(c *gin.Context) {
	rooms, err := rc.roomUsecase.GetAllPublicRooms(c)
	if err != nil {
		c.JSON(err.StatusCode, domain.ErrorResponse{
			Message: err.Message,
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Message: "Public rooms fetched successfully!",
		Data:    rooms,
	})
}

func (rc *RoomController) LeaveRoom(c *gin.Context) {
	err := rc.roomUsecase.LeaveRoom(c)

	if err != nil {
		c.JSON(err.StatusCode, domain.ErrorResponse{
			Message: err.Message,
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Message: "Left room successfully!",
	})
}
