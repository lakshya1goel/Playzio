package usecase

import (
	"fmt"
	"time"

	"github.com/lakshya1goel/Playzio/domain/model"
)

type GameUsecase interface {
	StartNextTurn()
	countAlivePlayers() int
	startTurn(userID uint)
	endGame(winnerID uint)
	checkEndCondition() bool
	getFinalScores() map[string]any
}

type gameUsecase struct {
	Pool          *model.GamePool
	GameRoomState *model.GameRoomState
}

func NewGameUsecase(pool *model.GamePool, room *model.GameRoomState) GameUsecase {
	return &gameUsecase{
		Pool:          pool,
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

	originalTurnIndex := g.GameRoomState.TurnIndex
	roundIncremented := false

	for i := 0; i < len(g.GameRoomState.Players); i++ {
		g.GameRoomState.TurnIndex = (g.GameRoomState.TurnIndex + 1) % len(g.GameRoomState.Players)
		uid := g.GameRoomState.Players[g.GameRoomState.TurnIndex]
		if g.GameRoomState.Lives[uid] > 0 {
			if !roundIncremented && g.GameRoomState.TurnIndex <= originalTurnIndex && originalTurnIndex != len(g.GameRoomState.Players)-1 {
				g.GameRoomState.Round++
				roundIncremented = true
			}
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
	g.GameRoomState.TimeLimit = 20 - g.GameRoomState.Round
	if g.GameRoomState.TimeLimit < 5 {
		g.GameRoomState.TimeLimit = 5
	}

	g.Pool.BroadcastToRoom(g.GameRoomState.RoomID, model.GameMessage{
		Type:   model.NextTurn,
		RoomID: g.GameRoomState.RoomID,
		Payload: map[string]any{
			"user_id":    userID,
			"char_set":   g.GameRoomState.CharSet,
			"time_limit": g.GameRoomState.TimeLimit,
			"round":      g.GameRoomState.Round,
		},
	})

	go func(uid uint, timeLimit int) {
		timer := time.NewTimer(time.Duration(timeLimit) * time.Second)
		defer timer.Stop()

		<-timer.C

		if g.GameRoomState.TurnIndex < len(g.GameRoomState.Players) &&
			g.GameRoomState.Players[g.GameRoomState.TurnIndex] == uid {

			g.GameRoomState.Lives[uid]--

			g.Pool.BroadcastToRoom(g.GameRoomState.RoomID, model.GameMessage{
				Type:   model.Timeout,
				RoomID: g.GameRoomState.RoomID,
				UserID: uid,
				Payload: map[string]any{
					"lives_left": g.GameRoomState.Lives[uid],
					"round":      g.GameRoomState.Round,
					"winner_id":  g.GameRoomState.WinnerID,
				},
			})
			g.StartNextTurn()
		}
	}(userID, g.GameRoomState.TimeLimit)
}

func (g *gameUsecase) endGame(winnerID uint) {
	g.GameRoomState.Started = false
	g.GameRoomState.WinnerID = winnerID

	g.Pool.BroadcastToRoom(g.GameRoomState.RoomID, model.GameMessage{
		Type:   model.GameOver,
		RoomID: g.GameRoomState.RoomID,
		Payload: map[string]any{
			"winner_id":    winnerID,
			"final_scores": g.getFinalScores(),
		},
	})
}

func (g *gameUsecase) checkEndCondition() bool {
	aliveCount := 0
	var lastAlivePlayer uint
	var highestScorePlayer uint
	maxScore := -1

	for uid, life := range g.GameRoomState.Lives {
		if life > 0 {
			aliveCount++
			lastAlivePlayer = uid
		}

		if g.GameRoomState.Points[uid] > maxScore {
			maxScore = g.GameRoomState.Points[uid]
			highestScorePlayer = uid
		}
	}

	if aliveCount <= 1 {
		var winnerID uint
		if aliveCount == 1 {
			winnerID = lastAlivePlayer
		} else {
			winnerID = highestScorePlayer
		}
		g.endGame(winnerID)
		return true
	}
	return false
}

func (g *gameUsecase) getFinalScores() map[string]any {
	scores := make(map[string]any)
	for uid, points := range g.GameRoomState.Points {
		scores[fmt.Sprintf("user_%d", uid)] = map[string]any{
			"points": points,
			"lives":  g.GameRoomState.Lives[uid],
		}
	}
	return scores
}
