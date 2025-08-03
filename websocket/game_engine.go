package websocket

import (
	"fmt"
	"time"

	"github.com/lakshya1goel/Playzio/bootstrap/util"
	"github.com/lakshya1goel/Playzio/domain/model"
)

type GameEngine interface {
	StartNextTurn()
	countAlivePlayers() int
	startTurn(userID uint)
	endGame(winnerID uint)
	checkEndCondition() bool
	getFinalScores() map[string]any
	handleSuccessfulAnswer(userID uint, answer string, newCharSet string)
	handleWrongAnswer(userID uint, answer string)
}

type gameEngine struct {
	Pool              *GamePool
	GameRoomState     *model.GameRoomState
	RoundMaxTimeLimit int
	MinTimeLimit      int
}

func NewGameEngine(pool *GamePool, room *model.GameRoomState) GameEngine {
	return &gameEngine{
		Pool:              pool,
		GameRoomState:     room,
		RoundMaxTimeLimit: RoundMaxTimeLimit,
		MinTimeLimit:      MinTimeLimit,
	}
}

func (g *gameEngine) StartNextTurn() {
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

func (g *gameEngine) countAlivePlayers() int {
	count := 0
	for _, life := range g.GameRoomState.Lives {
		if life > 0 {
			count++
		}
	}
	return count
}

func (g *gameEngine) startTurn(userID uint) {
	g.GameRoomState.TimeLimit = max(g.RoundMaxTimeLimit-g.GameRoomState.Round, g.MinTimeLimit)

	currentTurnIndex := g.GameRoomState.TurnIndex
	newCharSet := util.GenerateRandomWord()

	message := NewGameMessage().
		SetMessageType(model.NextTurn).
		WithRoomId(g.GameRoomState.RoomID).
		WithUserId(userID).
		WithCharSet(newCharSet).
		WithTimeLimit(g.GameRoomState.TimeLimit).
		WithRound(g.GameRoomState.Round).
		WithLives(g.GameRoomState.Lives[userID]).
		Build()

	g.Pool.BroadcastToRoom(g.GameRoomState.RoomID, message)

	go func(uid uint, timeLimit int, turnIndex int) {
		timer := time.NewTimer(time.Duration(timeLimit) * time.Second)
		defer timer.Stop()

		<-timer.C

		if g.GameRoomState.TurnIndex == turnIndex &&
			g.GameRoomState.Players[g.GameRoomState.TurnIndex] == uid {

			g.GameRoomState.Lives[uid]--

			message := NewGameMessage().
				SetMessageType(model.TurnEnded).
				WithRoomId(g.GameRoomState.RoomID).
				WithUserId(uid).
				WithReason("timeout").
				WithLives(g.GameRoomState.Lives[uid]).
				WithRound(g.GameRoomState.Round).
				WithScore(g.GameRoomState.Points[uid]).
				Build()

			g.Pool.BroadcastToRoom(g.GameRoomState.RoomID, message)

			g.StartNextTurn()
		}
	}(userID, g.GameRoomState.TimeLimit, currentTurnIndex)
}

func (g *gameEngine) handleSuccessfulAnswer(userID uint, answer string, newCharSet string) {
	message := NewGameMessage().
		SetMessageType(model.TurnEnded).
		WithRoomId(g.GameRoomState.RoomID).
		WithUserId(userID).
		WithReason("correct_answer").
		WithLives(g.GameRoomState.Lives[userID]).
		WithRound(g.GameRoomState.Round).
		WithScore(g.GameRoomState.Points[userID]).
		Build()

	g.Pool.BroadcastToRoom(g.GameRoomState.RoomID, message)

	g.StartNextTurn()
}

func (g *gameEngine) handleWrongAnswer(userID uint, answer string) {
	message := NewGameMessage().
		SetMessageType(model.TurnEnded).
		WithRoomId(g.GameRoomState.RoomID).
		WithUserId(userID).
		WithReason("wrong_answer").
		WithLives(g.GameRoomState.Lives[userID]).
		WithRound(g.GameRoomState.Round).
		WithScore(g.GameRoomState.Points[userID]).
		Build()

	g.Pool.BroadcastToRoom(g.GameRoomState.RoomID, message)

	if g.checkEndCondition() {
		return
	}

	if g.GameRoomState.Lives[userID] == 0 &&
		g.GameRoomState.Players[g.GameRoomState.TurnIndex] == userID {
		g.StartNextTurn()
	}
}

func (g *gameEngine) endGame(winnerID uint) {
	g.GameRoomState.Started = false
	g.GameRoomState.WinnerID = winnerID

	message := NewGameMessage().
		SetMessageType(model.GameOver).
		WithRoomId(g.GameRoomState.RoomID).
		WithWinnerId(winnerID).
		WithFinalScores(g.getFinalScores()).
		Build()

	g.Pool.BroadcastToRoom(g.GameRoomState.RoomID, message)
}

func (g *gameEngine) checkEndCondition() bool {
	aliveCount := 0
	var lastAlivePlayer uint
	var highestScorePlayer uint
	maxScore := MaxScoreForComparison

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

	if aliveCount <= MinAlivePlayersForGameEnd {
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

func (g *gameEngine) getFinalScores() map[string]any {
	scores := make(map[string]any)
	for uid, points := range g.GameRoomState.Points {
		scores[fmt.Sprintf("user_%d", uid)] = map[string]any{
			"points": points,
			"lives":  g.GameRoomState.Lives[uid],
		}
	}
	return scores
}
