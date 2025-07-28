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
	startRoomCountdown(roomID uint, gameRoomState *model.GameRoomState)
	handleCountdownEnd(roomID uint)
	BroadcastTimerStarted(roomID uint, duration int)
	addPlayerToRoom(client *GameClient, gameRoomState *model.GameRoomState)
	cleanupEmptyRoom(roomID uint)
	BroadcastToRoom(roomID uint, msg model.GameMessage)
	isRoomFull(roomID uint) bool
	isPlayerInRoom(playerID uint, gameRoomState *model.GameRoomState) bool
	canStartGame(gameRoomState *model.GameRoomState) bool
	isGameActive(gameRoomState *model.GameRoomState) bool
	JoinRoom(c *GameClient, roomID uint)
	RoomCount(roomID uint) int
	GetRemainingCountdownTime(roomID uint) int
	validateAnswer(answer string, gameRoomState *model.GameRoomState) bool
	processCorrectAnswer(c *GameClient, answer string, gameRoomState *model.GameRoomState)
	processWrongAnswer(c *GameClient, answer string, gameRoomState *model.GameRoomState)
	BroadcastMessage(c *GameClient, msg model.GameMessage)
	extractStringFromPayload(msg model.GameMessage, key string) (string, bool)
	extractUintFromPayload(msg model.GameMessage, key string) (uint, bool)
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
		RoundMaxTimeLimit: DefaultRoundTimeLimit,
		MinTimeLimit:      DefaultMinTimeLimit,
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

	g.BroadcastToRoom(g.GameRoomState.RoomID, model.NewNextTurnMessage(g.GameRoomState.RoomID, userID, newCharSet, g.GameRoomState.TimeLimit, g.GameRoomState.Round))

	go func(uid uint, timeLimit int, turnIndex int) {
		timer := time.NewTimer(time.Duration(timeLimit) * time.Second)
		defer timer.Stop()

		<-timer.C

		if g.GameRoomState.TurnIndex == turnIndex &&
			g.GameRoomState.Players[g.GameRoomState.TurnIndex] == uid {

			g.GameRoomState.Lives[uid]--
			turnEndedMsg := model.NewTurnEndedMessage(g.GameRoomState.RoomID, uid, "timeout", g.GameRoomState.Lives[uid], g.GameRoomState.Round, g.GameRoomState.Points[uid])
			g.BroadcastToRoom(g.GameRoomState.RoomID, turnEndedMsg)

			g.StartNextTurn()
		}
	}(userID, g.GameRoomState.TimeLimit, currentTurnIndex)
}

func (g *gameUsecase) endGame(winnerID uint) {
	g.GameRoomState.Started = false
	g.GameRoomState.WinnerID = winnerID

	gameOverMsg := model.NewGameOverMessage(g.GameRoomState.RoomID, winnerID, g.getFinalScores())
	g.BroadcastToRoom(g.GameRoomState.RoomID, gameOverMsg)
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

func (g *gameUsecase) handleSuccessfulAnswer(userID uint, answer string, newCharSet string) {
	turnEndedMsg := model.NewTurnEndedMessage(g.GameRoomState.RoomID, userID, "correct_answer", g.GameRoomState.Lives[userID], g.GameRoomState.Round, g.GameRoomState.Points[userID])
	g.BroadcastToRoom(g.GameRoomState.RoomID, turnEndedMsg)

	g.StartNextTurn()
}

func (g *gameUsecase) handleWrongAnswer(userID uint, answer string) {
	turnEndedMsg := model.NewTurnEndedMessage(g.GameRoomState.RoomID, userID, "wrong_answer", g.GameRoomState.Lives[userID], g.GameRoomState.Round, g.GameRoomState.Points[userID])
	g.BroadcastToRoom(g.GameRoomState.RoomID, turnEndedMsg)

	if g.checkEndCondition() {
		return
	}

	if g.GameRoomState.Lives[userID] == 0 &&
		g.GameRoomState.Players[g.GameRoomState.TurnIndex] == userID {
		g.StartNextTurn()
	}
}

func (g *gameUsecase) startRoomCountdown(roomID uint, gameRoomState *model.GameRoomState) {
	gameRoomState.CountdownTimer = time.AfterFunc(DefaultCountdownTime, func() {
		g.handleCountdownEnd(roomID)
	})
	g.BroadcastTimerStarted(roomID, DefaultCountdownSecs)
}

func (g *gameUsecase) handleCountdownEnd(roomID uint) {
	g.Pool.stateMu.Lock()
	defer g.Pool.stateMu.Unlock()

	gameRoomState := g.Pool.RoomsState[roomID]
	if !g.canStartGame(gameRoomState) {
		return
	}

	gameRoomState.Started = true
	gameRoomState.CharSet = util.GenerateRandomWord()
	gameRoomState.Round = 1
	gameRoomState.TimeLimit = DefaultGameTimeLimit
	gameRoomState.CountdownStarted = false
	gameRoomState.TurnIndex = 0

	startGameMsg := model.NewStartGameMessage(roomID, gameRoomState.CharSet, gameRoomState.Round, gameRoomState.TimeLimit)
	g.BroadcastToRoom(roomID, startGameMsg)

	game := NewGameUsecase(g.Pool, gameRoomState)
	game.StartNextTurn()
}

func (g *gameUsecase) BroadcastTimerStarted(roomID uint, duration int) {
	g.Pool.mu.RLock()
	clients := g.Pool.Rooms[roomID]
	if len(clients) == 0 {
		g.Pool.mu.RUnlock()
		return
	}
	for _, client := range clients {
		timerStartedMsg := model.NewTimerStartedMessage(roomID, duration)
		go client.WriteJSON(timerStartedMsg)
	}
	g.Pool.mu.RUnlock()
}

func (g *gameUsecase) addPlayerToRoom(client *GameClient, gameRoomState *model.GameRoomState) {
	if !g.isPlayerInRoom(client.UserId, gameRoomState) {
		gameRoomState.Players = append(gameRoomState.Players, client.UserId)
		gameRoomState.Lives[client.UserId] = DefaultPlayerLives
		gameRoomState.Points[client.UserId] = DefaultPlayerPoints

		if gameRoomState.CountdownStarted && !gameRoomState.Started {
			remainingTime := int(time.Until(gameRoomState.CountdownEndTime).Seconds())
			if remainingTime > 0 {
				g.BroadcastTimerStarted(client.RoomID, remainingTime)
			}
		}

		userJoinedMsg := model.NewUserJoinedMessage(client.UserId, client.UserName, client.RoomID)
		g.BroadcastToRoom(client.RoomID, userJoinedMsg)
	}
}

func (g *gameUsecase) cleanupEmptyRoom(roomID uint) {
	g.Pool.stateMu.Lock()
	defer g.Pool.stateMu.Unlock()

	if len(g.Pool.Rooms[roomID]) == 0 {
		if gameRoomState := g.Pool.RoomsState[roomID]; gameRoomState != nil && gameRoomState.CountdownTimer != nil {
			gameRoomState.CountdownTimer.Stop()
		}
		delete(g.Pool.RoomsState, roomID)
	}
}

func (g *gameUsecase) BroadcastToRoom(roomID uint, msg model.GameMessage) {
	g.Pool.mu.RLock()
	defer g.Pool.mu.RUnlock()
	for _, client := range g.Pool.Rooms[roomID] {
		go client.WriteJSON(msg)
	}
}

func (g *gameUsecase) isRoomFull(roomID uint) bool {
	return g.RoomCount(roomID) >= DefaultMaxPlayers
}

func (g *gameUsecase) isPlayerInRoom(playerID uint, gameRoomState *model.GameRoomState) bool {
	_, exists := gameRoomState.Lives[playerID]
	return exists
}

func (g *gameUsecase) canStartGame(gameRoomState *model.GameRoomState) bool {
	return gameRoomState != nil && !gameRoomState.Started
}

func (g *gameUsecase) isGameActive(gameRoomState *model.GameRoomState) bool {
	return gameRoomState != nil && gameRoomState.Started
}

func (g *gameUsecase) JoinRoom(c *GameClient, roomID uint) {
	// c.RoomID = roomID
	if g.isRoomFull(roomID) {
		fmt.Println("Room is full, cannot join:", roomID)
		return
	}
	g.Pool.Register <- c

	remainingTime := g.GetRemainingCountdownTime(roomID)
	if remainingTime > 0 {
		timerStartedMsg := model.NewTimerStartedMessage(roomID, remainingTime)
		c.WriteJSON(timerStartedMsg)
	}
}

func (g *gameUsecase) RoomCount(roomID uint) int {
	g.Pool.mu.RLock()
	defer g.Pool.mu.RUnlock()

	return len(g.Pool.Rooms[roomID])
}

func (g *gameUsecase) GetRemainingCountdownTime(roomID uint) int {
	g.Pool.stateMu.RLock()
	defer g.Pool.stateMu.RUnlock()

	gameRoomState := g.Pool.RoomsState[roomID]
	if gameRoomState == nil || !gameRoomState.CountdownStarted || gameRoomState.Started {
		return 0
	}

	remainingTime := int(time.Until(gameRoomState.CountdownEndTime).Seconds())
	if remainingTime < 0 {
		return 0
	}
	return remainingTime
}

func (g *gameUsecase) validateAnswer(answer string, gameRoomState *model.GameRoomState) bool {
	return util.ContainsSubstring(answer, gameRoomState.CharSet) && util.IsWordValid(answer)
}

func (g *gameUsecase) processCorrectAnswer(c *GameClient, answer string, gameRoomState *model.GameRoomState) {
	gameRoomState.CharSet = util.GenerateRandomWord()
	gameRoomState.Points[c.UserId]++

	responseMsg := model.NewAnswerResponseMessage(
		true,
		answer,
		c.RoomID,
		c.UserId,
		gameRoomState.CharSet,
		gameRoomState.Points[c.UserId],
		gameRoomState.Lives[c.UserId],
	)
	g.BroadcastMessage(c, responseMsg)

	g.handleSuccessfulAnswer(c.UserId, answer, gameRoomState.CharSet)
}

func (g *gameUsecase) processWrongAnswer(c *GameClient, answer string, gameRoomState *model.GameRoomState) {
	responseMsg := model.NewAnswerResponseMessage(
		false,
		answer,
		c.RoomID,
		c.UserId,
		"",
		0,
		gameRoomState.Lives[c.UserId],
	)
	g.BroadcastMessage(c, responseMsg)

	g.handleWrongAnswer(c.UserId, answer)
}

func (g *gameUsecase) BroadcastMessage(c *GameClient, msg model.GameMessage) {
	msg.Payload.(map[string]any)["user_id"] = c.UserId
	msg.Payload.(map[string]any)["room_id"] = c.RoomID
	g.Pool.Broadcast <- msg
}

func (g *gameUsecase) extractStringFromPayload(msg model.GameMessage, key string) (string, bool) {
	payload, ok := msg.Payload.(map[string]any)
	if !ok {
		return "", false
	}

	value, ok := payload[key]
	if !ok {
		return "", false
	}

	strValue, ok := value.(string)
	return strValue, ok
}

func (g *gameUsecase) extractUintFromPayload(msg model.GameMessage, key string) (uint, bool) {
	payload, ok := msg.Payload.(map[string]any)
	if !ok {
		fmt.Println("Invalid payload type")
		return 0, false
	}

	value, ok := payload[key]
	if !ok {
		fmt.Println("Key not found in payload")
		return 0, false
	}

	switch v := value.(type) {
	case float64:
		return uint(v), true
	case uint:
		return v, true
	case int:
		return uint(v), true
	default:
		fmt.Println("Invalid value type")
		return 0, false
	}
}
