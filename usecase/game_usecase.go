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
	GameRoomState *model.GameRoomState
}

func NewGameUsecase(pool *model.GamePool, room *model.GameRoomState) GameUsecase {
	return &gameUsecase{
		Pool: pool,
		GameRoomState: room,
	}
}

func (g *gameUsecase) StartNextTurn() {
	if !g.GameRoomState.Started {
		return
	}

	if g.checkEndCondition() {
		return
	}

	for i := 0; i < len(g.GameRoomState.Players); i++ {
		g.GameRoomState.TurnIndex = (g.GameRoomState.TurnIndex + 1) % len(g.GameRoomState.Players)
		uid := g.GameRoomState.Players[g.GameRoomState.TurnIndex]
		if g.GameRoomState.Lives[uid] > 0 {
			g.startTurn(uid)
			break
		}
	}
}

func (g *gameUsecase) countAlivePlayers() int {
	count := 0
	for _, life := range g.GameRoomState.Lives {
		if life > 0 {
			count++
		}
	}
	return count
}

func (g *gameUsecase) startTurn(userID uint) {
	turnNum := g.GameRoomState.TurnIndex + 1
	timeLimit := 12 - turnNum
	if timeLimit < 8 {
		timeLimit = 8
	}

	g.Pool.BroadcastToRoom(g.GameRoomState.RoomID, model.GameMessage{
		Type:   model.NextTurn,
		RoomID: g.GameRoomState.RoomID,
		Payload: map[string]any{
			"user_id":    userID,
			"char_set":   g.GameRoomState.CharSet,
			"time_limit": timeLimit,
		},
	})

	go func(uid uint) {
		timer := time.NewTimer(time.Duration(timeLimit) * time.Second)
		defer timer.Stop()

		<-timer.C

		if g.GameRoomState.TurnIndex < len(g.GameRoomState.Players) && g.GameRoomState.Players[g.GameRoomState.TurnIndex] == uid {
			g.GameRoomState.Lives[uid]--
			g.Pool.BroadcastToRoom(g.GameRoomState.RoomID, model.GameMessage{
				Type:   model.Timeout,
				RoomID: g.GameRoomState.RoomID,
				UserID: uid,
				Payload: map[string]any{
					"lives_left": g.GameRoomState.Lives[uid],
				},
			})
			g.StartNextTurn()
		}
	}(userID)
}

func (g *gameUsecase) endGame(winnerID uint) {
	g.GameRoomState.Started = false

	g.Pool.BroadcastToRoom(g.GameRoomState.RoomID, model.GameMessage{
		Type:   model.GameOver,
		RoomID: g.GameRoomState.RoomID,
		Payload: map[string]any{
			"winner_id": winnerID,
		},
	})
}

func (g *gameUsecase) checkEndCondition() bool {
	aliveCount := 0
	var lastAlivePlayer uint

	for uid, life := range g.GameRoomState.Lives {
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
