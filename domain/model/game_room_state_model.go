package model

type GameRoomState struct {
	RoomID    uint
	CreatedBy uint
	Players   []uint
	Lives     map[uint]int
	Points    map[uint]int
	TurnIndex int
	CharSet   string
	Started   bool
	Capacity  int
}
