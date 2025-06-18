package usecase

import "github.com/lakshya1goel/Playzio/repository"

type RoomUsecase interface {

}

type roomUsecase struct {
	roomRepo repository.RoomRepository
}

func NewRoomUsecase() RoomUsecase {
	return &roomUsecase{
		roomRepo: repository.NewRoomRepository(),
	}
}