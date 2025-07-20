package websocket

import (
	"fmt"
	"sync"
	"time"

	"github.com/lakshya1goel/Playzio/bootstrap/util"
	"github.com/lakshya1goel/Playzio/domain/model"
)

const (
	DefaultPlayerLives    = 3
	DefaultPlayerPoints   = 0
	DefaultMaxPlayers     = 10
	DefaultCountdownTime  = 2 * time.Minute
	DefaultCountdownSecs  = 120
	DefaultGameTimeLimit  = 19
	DefaultRoundTimeLimit = 20
	DefaultMinTimeLimit   = 5
)

type GamePool struct {
	*BasePool[*GameClient]
	RoomsState map[uint]*model.GameRoomState
	stateMu    sync.RWMutex
}

func NewGamePool() *GamePool {
	return &GamePool{
		BasePool:   NewBasePool[*GameClient](),
		RoomsState: make(map[uint]*model.GameRoomState),
	}
}

func (p *GamePool) Start() {
	for {
		select {
		case client := <-p.Register:
			util.RegisterClient(&p.mu, p.Rooms, client.RoomID, client.UserId, client)

			p.stateMu.Lock()
			if _, ok := p.RoomsState[client.RoomID]; !ok {
				gameRoomState := p.initializeGameRoomState(client.RoomID, client.UserId)
				p.RoomsState[client.RoomID] = gameRoomState
				p.startRoomCountdown(client.RoomID, gameRoomState)
				userJoinedMsg := model.NewUserJoinedMessage(client.UserId, client.UserName, "Room created and joined", client.RoomID)
				p.BroadcastToRoom(client.RoomID, userJoinedMsg)
			} else {
				p.addPlayerToRoom(client, p.RoomsState[client.RoomID])
			}
			p.stateMu.Unlock()

		case client := <-p.Unregister:
			userLeftMsg := model.NewUserLeftMessage(client.UserId, client.UserName, "User left the room", client.RoomID)
			p.BroadcastToRoom(client.RoomID, userLeftMsg)

			util.UnregisterClient(&p.mu, p.Rooms, client.RoomID, client.UserId)
			p.cleanupEmptyRoom(client.RoomID)

		case raw := <-p.Broadcast:
			msg, ok := raw.(model.GameMessage)
			if !ok {
				continue
			}
			p.mu.RLock()
			clients := p.Rooms[msg.Payload.(map[string]any)["room_id"].(uint)]
			for _, client := range clients {
				go client.WriteJSON(msg)
			}
			p.mu.RUnlock()
		}
	}
}

func (p *GamePool) initializeGameRoomState(roomID uint, creatorID uint) *model.GameRoomState {
	return &model.GameRoomState{
		RoomID:           roomID,
		CreatedBy:        creatorID,
		Players:          []uint{creatorID},
		Lives:            map[uint]int{creatorID: DefaultPlayerLives},
		Points:           map[uint]int{creatorID: DefaultPlayerPoints},
		TurnIndex:        0,
		CharSet:          "",
		Started:          false,
		Round:            0,
		TimeLimit:        0,
		WinnerID:         0,
		CountdownStarted: true,
		CountdownEndTime: time.Now().Add(DefaultCountdownTime),
	}
}

func (p *GamePool) startRoomCountdown(roomID uint, gameRoomState *model.GameRoomState) {
	gameRoomState.CountdownTimer = time.AfterFunc(DefaultCountdownTime, func() {
		p.handleCountdownEnd(roomID)
	})
	p.BroadcastTimerStarted(roomID, DefaultCountdownSecs)
}

func (p *GamePool) handleCountdownEnd(roomID uint) {
	p.stateMu.Lock()
	defer p.stateMu.Unlock()

	gameRoomState := p.RoomsState[roomID]
	if !p.canStartGame(gameRoomState) {
		return
	}

	gameRoomState.Started = true
	gameRoomState.CharSet = util.GenerateRandomWord()
	gameRoomState.Round = 1
	gameRoomState.TimeLimit = DefaultGameTimeLimit
	gameRoomState.CountdownStarted = false
	gameRoomState.TurnIndex = 0

	startGameMsg := model.NewStartGameMessage(roomID, "Game has started", gameRoomState.CharSet, gameRoomState.Round, gameRoomState.TimeLimit)
	p.BroadcastToRoom(roomID, startGameMsg)

	game := NewGameUsecase(p, gameRoomState)
	game.StartNextTurn()
}

func (p *GamePool) BroadcastTimerStarted(roomID uint, duration int) {
	p.mu.RLock()
	clients := p.Rooms[roomID]
	if len(clients) == 0 {
		p.mu.RUnlock()
		return
	}
	for _, client := range clients {
		timerStartedMsg := model.NewTimerStartedMessage(roomID, duration)
		go client.WriteJSON(timerStartedMsg)
	}
	p.mu.RUnlock()
}

func (p *GamePool) addPlayerToRoom(client *GameClient, gameRoomState *model.GameRoomState) {
	if !p.isPlayerInRoom(client.UserId, gameRoomState) {
		gameRoomState.Players = append(gameRoomState.Players, client.UserId)
		gameRoomState.Lives[client.UserId] = DefaultPlayerLives
		gameRoomState.Points[client.UserId] = DefaultPlayerPoints

		if gameRoomState.CountdownStarted && !gameRoomState.Started {
			remainingTime := int(time.Until(gameRoomState.CountdownEndTime).Seconds())
			if remainingTime > 0 {
				p.BroadcastTimerStarted(client.RoomID, remainingTime)
			}
		}

		userJoinedMsg := model.NewUserJoinedMessage(client.UserId, client.UserName, "User joined the room", client.RoomID)
		p.BroadcastToRoom(client.RoomID, userJoinedMsg)
	}
}

func (p *GamePool) cleanupEmptyRoom(roomID uint) {
	p.stateMu.Lock()
	defer p.stateMu.Unlock()

	if len(p.Rooms[roomID]) == 0 {
		if gameRoomState := p.RoomsState[roomID]; gameRoomState != nil && gameRoomState.CountdownTimer != nil {
			gameRoomState.CountdownTimer.Stop()
		}
		delete(p.RoomsState, roomID)
	}
}

func (p *GamePool) BroadcastToRoom(roomID uint, msg model.GameMessage) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, client := range p.Rooms[roomID] {
		go client.WriteJSON(msg)
	}
}

func (p *GamePool) Read(c *GameClient) {
	defer func() {
		c.StopPingPong()
		p.LeaveRoom(c)
		c.Conn.Close()
	}()

	c.StartPingPong()

	for {
		var msg model.GameMessage
		err := c.Conn.ReadJSON(&msg)

		if err != nil {
			fmt.Println("Game WebSocket read error:", err)
			return
		}

		switch msg.Type {
		case model.Join:
			if !p.handleJoinMessage(c, msg) {
				continue
			}
		case model.Answer:
			if !p.handleAnswerMessage(c, msg) {
				continue
			}
		case model.Leave:
			p.LeaveRoom(c)
		case model.Typing:
			if !p.handleTypingMessage(c, msg) {
				continue
			}
		case model.Ping:
			p.handlePingMessage(c, msg)
		case model.Pong:
			p.handlePongMessage(c, msg)
		default:
			fmt.Println("Unknown game message type:", msg.Type)
		}
	}
}

//handleJoinMessage handles the join message
func (p *GamePool) handleJoinMessage(c *GameClient, msg model.GameMessage) bool {
	roomID, ok := p.extractUintFromPayload(msg, "room_id")
	if !ok || roomID == 0 {
		fmt.Println("Invalid Room ID received")
		return false
	}
	p.JoinRoom(c, roomID)
	return true
}

func (p *GamePool) isRoomFull(roomID uint) bool {
	return p.RoomCount(roomID) >= DefaultMaxPlayers
}

func (p *GamePool) isPlayerInRoom(playerID uint, gameRoomState *model.GameRoomState) bool {
	_, exists := gameRoomState.Lives[playerID]
	return exists
}

func (p *GamePool) canStartGame(gameRoomState *model.GameRoomState) bool {
	return gameRoomState != nil && !gameRoomState.Started
}

func (p *GamePool) isGameActive(gameRoomState *model.GameRoomState) bool {
	return gameRoomState != nil && gameRoomState.Started
}

func (p *GamePool) JoinRoom(c *GameClient, roomID uint) {
	c.RoomID = roomID
	if p.isRoomFull(roomID) {
		fmt.Println("Room is full, cannot join:", roomID)
		return
	}
	p.Register <- c

	remainingTime := p.GetRemainingCountdownTime(roomID)
	if remainingTime > 0 {
		timerStartedMsg := model.NewTimerStartedMessage(roomID, remainingTime)
		c.WriteJSON(timerStartedMsg)
	}
}

func (p *BasePool[T]) RoomCount(roomID uint) int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.Rooms[roomID])
}

func (p *GamePool) GetRemainingCountdownTime(roomID uint) int {
	p.stateMu.RLock()
	defer p.stateMu.RUnlock()

	gameRoomState := p.RoomsState[roomID]
	if gameRoomState == nil || !gameRoomState.CountdownStarted || gameRoomState.Started {
		return 0
	}

	remainingTime := int(time.Until(gameRoomState.CountdownEndTime).Seconds())
	if remainingTime < 0 {
		return 0
	}
	return remainingTime
}

//handleAnswerMessage handles the answer message
func (p *GamePool) handleAnswerMessage(c *GameClient, msg model.GameMessage) bool {
	gameRoomState := p.RoomsState[c.RoomID]
	if !p.isGameActive(gameRoomState) {
		fmt.Println("Room not found or game not started")
		return false
	}

	answer, ok := p.extractStringFromPayload(msg, "answer")
	if !ok {
		fmt.Println("Answer not provided in payload")
		return false
	}

	if p.validateAnswer(answer, gameRoomState) {
		p.processCorrectAnswer(c, answer, gameRoomState)
	} else {
		p.processWrongAnswer(c, answer, gameRoomState)
	}
	return true
}

func (p *GamePool) validateAnswer(answer string, gameRoomState *model.GameRoomState) bool {
	return util.ContainsSubstring(answer, gameRoomState.CharSet) && util.IsWordValid(answer)
}

func (p *GamePool) processCorrectAnswer(c *GameClient, answer string, gameRoomState *model.GameRoomState) {
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
	p.BroadcastMessage(c, responseMsg)

	game := NewGameUsecase(c.Pool, gameRoomState)
	game.handleSuccessfulAnswer(c.UserId, answer, gameRoomState.CharSet)
}

func (p *GamePool) processWrongAnswer(c *GameClient, answer string, gameRoomState *model.GameRoomState) {
	responseMsg := model.NewAnswerResponseMessage(
		false,
		answer,
		c.RoomID,
		c.UserId,
		"",
		0,
		gameRoomState.Lives[c.UserId],
	)
	p.BroadcastMessage(c, responseMsg)

	game := NewGameUsecase(c.Pool, gameRoomState)
	game.handleWrongAnswer(c.UserId, answer)
}

//LeaveRoom leaves the room
func (p *GamePool) LeaveRoom(c *GameClient) {
	p.Unregister <- c
}

//handleTypingMessage handles the typing message
func (p *GamePool) handleTypingMessage(c *GameClient, msg model.GameMessage) bool {
	gameRoomState := c.Pool.RoomsState[c.RoomID]
	if gameRoomState == nil {
		return false
	}

	text, ok := p.extractStringFromPayload(msg, "text")
	if !ok {
		fmt.Println("Text not provided in typing payload")
		return false
	}

	typingMsg := model.NewTypingMessage(text, c.RoomID, c.UserId)
	p.Broadcast <- typingMsg
	return true
}

//handlePingMessage handles the ping message
func (p *GamePool) handlePingMessage(c *GameClient, msg model.GameMessage) {
	if payload, ok := msg.Payload.(map[string]any); ok {
		if timestamp, ok := payload["timestamp"].(float64); ok {
			c.SendPong(int64(timestamp))
		}
	}
}

//handlePongMessage handles the pong message
func (p *GamePool) handlePongMessage(c *GameClient, msg model.GameMessage) {
	if payload, ok := msg.Payload.(map[string]any); ok {
		if timestamp, ok := payload["timestamp"].(float64); ok {
			c.HandlePong(int64(timestamp))
		}
	}
}

//BroadcastMessage broadcasts a message to the room
func (p *GamePool) BroadcastMessage(c *GameClient, msg model.GameMessage) {
	msg.Payload.(map[string]any)["user_id"] = c.UserId
	msg.Payload.(map[string]any)["room_id"] = c.RoomID
	p.Broadcast <- msg
}

func (p *GamePool) extractStringFromPayload(msg model.GameMessage, key string) (string, bool) {
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

func (p *GamePool) extractUintFromPayload(msg model.GameMessage, key string) (uint, bool) {
	payload, ok := msg.Payload.(map[string]any)
	if !ok {
		return 0, false
	}

	value, ok := payload[key]
	if !ok {
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
		return 0, false
	}
}
