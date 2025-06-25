package model

type GameRoomState struct {
	RoomID    uint
	Players   []uint
	Lives     map[uint]int
	Points    map[uint]int
	TurnIndex int
	CharSet   string
}
