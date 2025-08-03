package websocket

import "time"

const (
	MaxRoomCapacity = 10
)

const (
	CountdownDuration = 2 * time.Minute
	RoundMaxTimeLimit = 20
	MinTimeLimit      = 5
	InitialTimeLimit  = 19
)

const (
	InitialLives     = 3
	InitialRound     = 1
	InitialPoints    = 0
	InitialTurnIndex = 0
)

const (
	MinAlivePlayersForGameEnd = 1
	MaxScoreForComparison     = -1
)

const (
	PingInterval = 30 * time.Second
	PongTimeout  = 10 * time.Second
)
