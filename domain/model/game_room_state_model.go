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
	Round     int
	TimeLimit int
	WinnerID  uint
}
