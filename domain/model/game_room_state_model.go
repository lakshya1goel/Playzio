package model

import "time"

type GameRoomState struct {
	RoomID           uint
	CreatedBy        uint
	Players          []uint
	Lives            map[uint]int
	Points           map[uint]int
	TurnIndex        int
	CharSet          string
	Started          bool
	Round            int
	TimeLimit        int
	WinnerID         uint
	CountdownStarted bool
	CountdownEndTime time.Time
	CountdownTimer   *time.Timer
}
