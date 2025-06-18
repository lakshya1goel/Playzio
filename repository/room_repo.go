package repository

type RoomRepository interface {

}

type roomRepository struct {}

func NewRoomRepository() RoomRepository {
	return &roomRepository{}
}