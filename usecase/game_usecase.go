package usecase

import (
	"time"

	"github.com/lakshya1goel/Playzio/domain/model"
)

type GameUsecase interface {
	StartNextTurn()
	countAlivePlayers() int
	startTurn(userID uint)
	endGame(winnerID uint)
	checkEndCondition() bool
}

type gameUsecase struct {
	Pool *model.GamePool
	Room *model.GameRoomState
}

func NewGameUsecase(pool *model.GamePool, room *model.GameRoomState) GameUsecase {
	return &gameUsecase{
		Pool: pool,
		Room: room,
	}
}

func (g *gameUsecase) StartNextTurn() {
	if !g.Room.Started {
		return
	}

	if g.checkEndCondition() {
		return
	}

	for i := 0; i < len(g.Room.Players); i++ {
		g.Room.TurnIndex = (g.Room.TurnIndex + 1) % len(g.Room.Players)
		uid := g.Room.Players[g.Room.TurnIndex]
		if g.Room.Lives[uid] > 0 {
			g.startTurn(uid)
			break
		}
	}
}

func (g *gameUsecase) countAlivePlayers() int {
	count := 0
	for _, life := range g.Room.Lives {
		if life > 0 {
			count++
		}
	}
	return count
}

func (g *gameUsecase) startTurn(userID uint) {
	turnNum := g.Room.TurnIndex + 1
	timeLimit := 12 - turnNum
	if timeLimit < 8 {
		timeLimit = 8
	}

	g.Pool.BroadcastToRoom(g.Room.RoomID, model.GameMessage{
		Type:   model.NextTurn,
		RoomID: g.Room.RoomID,
		Payload: map[string]any{
			"user_id":    userID,
			"char_set":   g.Room.CharSet,
			"time_limit": timeLimit,
		},
	})

	go func(uid uint) {
		timer := time.NewTimer(time.Duration(timeLimit) * time.Second)
		defer timer.Stop()

		<-timer.C

		if g.Room.TurnIndex < len(g.Room.Players) && g.Room.Players[g.Room.TurnIndex] == uid {
			g.Room.Lives[uid]--
			g.Pool.BroadcastToRoom(g.Room.RoomID, model.GameMessage{
				Type:   model.Timeout,
				RoomID: g.Room.RoomID,
				UserID: uid,
				Payload: map[string]any{
					"lives_left": g.Room.Lives[uid],
				},
			})
			g.StartNextTurn()
		}
	}(userID)
}

func (g *gameUsecase) endGame(winnerID uint) {
	g.Room.Started = false

	g.Pool.BroadcastToRoom(g.Room.RoomID, model.GameMessage{
		Type:   model.GameOver,
		RoomID: g.Room.RoomID,
		Payload: map[string]any{
			"winner_id": winnerID,
		},
	})
}

func (g *gameUsecase) checkEndCondition() bool {
	aliveCount := 0
	var lastAlivePlayer uint

	for uid, life := range g.Room.Lives {
		if life > 0 {
			aliveCount++
			lastAlivePlayer = uid
		}
	}

	if aliveCount <= 1 {
		g.endGame(lastAlivePlayer)
		return true
	}
	return false
}
