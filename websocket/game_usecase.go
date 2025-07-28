package websocket

import (
	"fmt"
	"time"

	"github.com/lakshya1goel/Playzio/bootstrap/util"
	"github.com/lakshya1goel/Playzio/domain/model"
)

type GameUsecase interface {
	StartNextTurn()
	countAlivePlayers() int
	startTurn(userID uint)
	endGame(winnerID uint)
	checkEndCondition() bool
	getFinalScores() map[string]any
	handleSuccessfulAnswer(userID uint, answer string, newCharSet string)
	handleWrongAnswer(userID uint, answer string)
}

type gameUsecase struct {
	Pool              *GamePool
	GameRoomState     *model.GameRoomState
	RoundMaxTimeLimit int
	MinTimeLimit      int
}

func NewGameUsecase(pool *GamePool, room *model.GameRoomState) GameUsecase {
	return &gameUsecase{
		Pool:              pool,
		GameRoomState:     room,
		RoundMaxTimeLimit: 20,
		MinTimeLimit:      5,
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
	g.GameRoomState.TimeLimit = max(g.RoundMaxTimeLimit-g.GameRoomState.Round, g.MinTimeLimit)

	currentTurnIndex := g.GameRoomState.TurnIndex
	newCharSet := util.GenerateRandomWord()

	message := NewGameMessage().
		setMessageType(model.NextTurn).
		withRoomId(g.GameRoomState.RoomID).
		withUserId(userID).
		withCharSet(newCharSet).
		withTimeLimit(g.GameRoomState.TimeLimit).
		withRound(g.GameRoomState.Round).
		withLives(g.GameRoomState.Lives[userID]).
		build()

	g.Pool.BroadcastToRoom(g.GameRoomState.RoomID, message)

	go func(uid uint, timeLimit int, turnIndex int) {
		timer := time.NewTimer(time.Duration(timeLimit) * time.Second)
		defer timer.Stop()

		<-timer.C

		if g.GameRoomState.TurnIndex == turnIndex &&
			g.GameRoomState.Players[g.GameRoomState.TurnIndex] == uid {

			g.GameRoomState.Lives[uid]--

			message := NewGameMessage().
				setMessageType(model.TurnEnded).
				withRoomId(g.GameRoomState.RoomID).
				withUserId(uid).
				withReason("timeout").
				withLives(g.GameRoomState.Lives[uid]).
				withRound(g.GameRoomState.Round).
				withScore(g.GameRoomState.Points[uid]).
				build()

			g.Pool.BroadcastToRoom(g.GameRoomState.RoomID, message)

			g.StartNextTurn()
		}
	}(userID, g.GameRoomState.TimeLimit, currentTurnIndex)
}

func (g *gameUsecase) handleSuccessfulAnswer(userID uint, answer string, newCharSet string) {
	message := NewGameMessage().
		setMessageType(model.TurnEnded).
		withRoomId(g.GameRoomState.RoomID).
		withUserId(userID).
		withReason("correct_answer").
		withLives(g.GameRoomState.Lives[userID]).
		withRound(g.GameRoomState.Round).
		withScore(g.GameRoomState.Points[userID]).
		build()

	g.Pool.BroadcastToRoom(g.GameRoomState.RoomID, message)

	g.StartNextTurn()
}

func (g *gameUsecase) handleWrongAnswer(userID uint, answer string) {
	message := NewGameMessage().
		setMessageType(model.TurnEnded).
		withRoomId(g.GameRoomState.RoomID).
		withUserId(userID).
		withReason("wrong_answer").
		withLives(g.GameRoomState.Lives[userID]).
		withRound(g.GameRoomState.Round).
		withScore(g.GameRoomState.Points[userID]).
		build()

	g.Pool.BroadcastToRoom(g.GameRoomState.RoomID, message)

	if g.checkEndCondition() {
		return
	}

	if g.GameRoomState.Lives[userID] == 0 &&
		g.GameRoomState.Players[g.GameRoomState.TurnIndex] == userID {
		g.StartNextTurn()
	}
}

func (g *gameUsecase) endGame(winnerID uint) {
	g.GameRoomState.Started = false
	g.GameRoomState.WinnerID = winnerID

	message := NewGameMessage().
		setMessageType(model.GameOver).
		withRoomId(g.GameRoomState.RoomID).
		withWinnerId(winnerID).
		withFinalScores(g.getFinalScores()).
		build()

	g.Pool.BroadcastToRoom(g.GameRoomState.RoomID, message)
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
