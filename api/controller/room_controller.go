package controller

import "github.com/lakshya1goel/Playzio/usecase"

type RoomController struct {
	roomUsecase usecase.RoomUsecase
}

func NewRoomController() *RoomController {
	return &RoomController{
		roomUsecase: usecase.NewRoomUsecase(),
	}
}
