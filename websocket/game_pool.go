package websocket

import (
	"fmt"
	"time"

	"github.com/lakshya1goel/Playzio/bootstrap/util"
	"github.com/lakshya1goel/Playzio/domain/model"
)

type GamePool struct {
	*BasePool[*GameClient]
	gameStateManager   GameStateManager
	gameTimerManager   GameTimerManager
	gameMessageHandler GameMessageHandler
}

func NewGamePool() *GamePool {
	pool := &GamePool{
		BasePool: NewBasePool[*GameClient](),
	}
	pool.gameStateManager = NewGameStateManager()
	pool.gameTimerManager = NewGameTimerManager(pool)
	pool.gameMessageHandler = NewGameMessageHandler(pool)
	return pool
}

func (p *GamePool) Start() {
	for {
		select {
		case client := <-p.Register:
			p.handleClientRegister(client)
		case client := <-p.Unregister:
			p.handleClientUnregister(client)
		case raw := <-p.Broadcast:
			p.handleBroadcast(raw)
		}
	}
}

func (p *GamePool) handleClientRegister(client *GameClient) {
	util.RegisterClient(&p.mu, p.Rooms, client.RoomID, client.UserId, client)

	roomState := p.gameStateManager.GetRoomState(client.RoomID)
	if roomState == nil {
		p.gameStateManager.CreateRoomState(client.RoomID, client.UserId)
		p.gameTimerManager.StartCountdown(client.RoomID, 2*time.Minute)
	} else {
		if p.gameStateManager.AddPlayer(client.RoomID, client.UserId) {
			remainingTime := p.gameTimerManager.GetRemainingCountdownTime(client.RoomID)
			if remainingTime > 0 {
				p.BroadcastTimerStarted(client.RoomID, remainingTime)
			}
		}
	}

	message := NewGameMessage().
		SetMessageType(model.UserJoined).
		WithUserId(client.UserId).
		WithRoomId(client.RoomID).
		WithUserName(client.UserName).
		Build()

	p.BroadcastToRoom(client.RoomID, message)
}

func (p *GamePool) handleClientUnregister(client *GameClient) {
	message := NewGameMessage().
		SetMessageType(model.UserLeft).
		WithUserId(client.UserId).
		WithRoomId(client.RoomID).
		WithUserName(client.UserName).
		Build()

	p.BroadcastToRoom(client.RoomID, message)

	util.UnregisterClient(&p.mu, p.Rooms, client.RoomID, client.UserId)

	if len(p.Rooms[client.RoomID]) == 0 {
		p.gameTimerManager.StopCountdown(client.RoomID)
		p.gameStateManager.RemoveRoom(client.RoomID)
	}
}

func (p *GamePool) handleBroadcast(raw interface{}) {
	msg, ok := raw.(model.GameMessage)
	if !ok {
		return
	}

	roomID, err := p.gameMessageHandler.ExtractRoomID(msg)
	if err != nil {
		return
	}

	p.mu.RLock()
	clients := p.Rooms[roomID]
	for _, client := range clients {
		go client.WriteJSON(msg)
	}
	p.mu.RUnlock()
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
			if !p.gameMessageHandler.HandleJoin(c, msg) {
				continue
			}
		case model.Answer:
			if !p.gameMessageHandler.HandleAnswer(c, msg) {
				continue
			}
		case model.Leave:
			p.LeaveRoom(c)
		case model.Typing:
			if !p.gameMessageHandler.HandleTyping(c, msg) {
				continue
			}
		case model.Ping:
			if timestamp, ok := msg.Payload["timestamp"].(float64); ok {
				c.SendPong(int64(timestamp))
			}
		case model.Pong:
			if timestamp, ok := msg.Payload["timestamp"].(float64); ok {
				c.HandlePong(int64(timestamp))
			}
		default:
			fmt.Println("Unknown game message type:", msg.Type)
		}
	}
}

func (p *GamePool) JoinRoom(c *GameClient, roomID uint) {
	c.RoomID = roomID
	if p.RoomCount(roomID) >= 10 {
		fmt.Println("Room is full, cannot join:", roomID)
		return
	}
	p.Register <- c

	remainingTime := p.gameTimerManager.GetRemainingCountdownTime(roomID)
	if remainingTime > 0 {
		message := NewGameMessage().
			SetMessageType(model.TimerStarted).
			WithRoomId(roomID).
			WithDuration(remainingTime).
			Build()

		c.WriteJSON(message)
	}
}

func (p *GamePool) LeaveRoom(c *GameClient) {
	p.Unregister <- c
}

func (p *GamePool) BroadcastMessage(c *GameClient, msg model.GameMessage) {
	msg.Payload["user_id"] = c.UserId
	msg.Payload["room_id"] = c.RoomID
	p.Broadcast <- msg
}

func (p *GamePool) handleCountdownEnd(roomID uint) {
	gameRoomState := p.gameStateManager.GetRoomState(roomID)
	if gameRoomState == nil || gameRoomState.Started {
		return
	}

	gameRoomState.Started = true
	gameRoomState.CharSet = util.GenerateRandomWord()
	gameRoomState.Round = 1
	gameRoomState.TimeLimit = 19
	gameRoomState.CountdownStarted = false
	gameRoomState.TurnIndex = 0

	message := NewGameMessage().
		SetMessageType(model.StartGame).
		WithRoomId(roomID).
		WithCharSet(gameRoomState.CharSet).
		WithRound(gameRoomState.Round).
		WithTimeLimit(gameRoomState.TimeLimit).
		Build()

	p.BroadcastToRoom(roomID, message)

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
		message := NewGameMessage().
			SetMessageType(model.TimerStarted).
			WithRoomId(roomID).
			WithDuration(duration).
			Build()

		go client.WriteJSON(message)
	}
	p.mu.RUnlock()
}

func (p *BasePool[T]) RoomCount(roomID uint) int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.Rooms[roomID])
}

func (p *GamePool) BroadcastToRoom(roomID uint, msg model.GameMessage) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, client := range p.Rooms[roomID] {
		go client.WriteJSON(msg)
	}
}
