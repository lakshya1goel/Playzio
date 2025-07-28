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
	RoomsState   map[uint]*model.GameRoomState
	stateMu      sync.RWMutex
	GameUsecases map[uint]GameUsecase
}

func NewGamePool() *GamePool {
	return &GamePool{
		BasePool:     NewBasePool[*GameClient](),
		RoomsState:   make(map[uint]*model.GameRoomState),
		GameUsecases: make(map[uint]GameUsecase),
	}
}

func (p *GamePool) Start() {
	for {
		select {
		case client := <-p.Register:
			p.handleRegister(client)

		case client := <-p.Unregister:
			p.handleUnregister(client)

		case raw := <-p.Broadcast:
			p.handleBroadcast(raw)
		}
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

func (p *GamePool) handleRegister(client *GameClient) {
	// fmt.Println("Handling register message for client", client.UserId, "in room", client.RoomID)
	// util.RegisterClient(&p.mu, p.Rooms, client.RoomID, client.UserId, client)

	// p.stateMu.Lock()
	// if _, ok := p.RoomsState[client.RoomID]; !ok {
	// 	gameRoomState := &model.GameRoomState{
	// 		RoomID:           client.RoomID,
	// 		CreatedBy:        client.UserId,
	// 		Players:          []uint{client.UserId},
	// 		Lives:            map[uint]int{client.UserId: DefaultPlayerLives},
	// 		Points:           map[uint]int{client.UserId: DefaultPlayerPoints},
	// 		TurnIndex:        0,
	// 		CharSet:          "",
	// 		Started:          false,
	// 		Round:            0,
	// 		TimeLimit:        0,
	// 		WinnerID:         0,
	// 		CountdownStarted: true,
	// 		CountdownEndTime: time.Now().Add(DefaultCountdownTime),
	// 	}
	// 	p.RoomsState[client.RoomID] = gameRoomState
	// 	p.GameUsecases[client.RoomID] = NewGameUsecase(p, gameRoomState)
	// 	p.GameUsecases[client.RoomID].startRoomCountdown(client.RoomID, gameRoomState)
	// 	userJoinedMsg := model.NewUserJoinedMessage(client.UserId, client.UserName, client.RoomID)
	// 	p.GameUsecases[client.RoomID].BroadcastToRoom(client.RoomID, userJoinedMsg)
	// } else {
	// 	p.GameUsecases[client.RoomID].addPlayerToRoom(client, p.RoomsState[client.RoomID])
	// }
	// p.stateMu.Unlock()
}

func (p *GamePool) handleUnregister(client *GameClient) {
	fmt.Println("Unregistering client", client.UserId, "from room", client.RoomID)
	userLeftMsg := model.NewUserLeftMessage(client.UserId, client.UserName, client.RoomID)
	p.GameUsecases[client.RoomID].BroadcastToRoom(client.RoomID, userLeftMsg)

	util.UnregisterClient(&p.mu, p.Rooms, client.RoomID, client.UserId)
	p.GameUsecases[client.RoomID].cleanupEmptyRoom(client.RoomID)
}

func (p *GamePool) handleBroadcast(raw interface{}) {
	msg, ok := raw.(model.GameMessage)
	if !ok {
		return
	}
	p.mu.RLock()
	clients := p.Rooms[msg.Payload.(map[string]any)["room_id"].(uint)]
	for _, client := range clients {
		go client.WriteJSON(msg)
	}
	p.mu.RUnlock()
}

func (p *GamePool) handleJoinMessage(c *GameClient, msg model.GameMessage) bool {
	roomIDFloat := msg.Payload.(map[string]any)["room_id"].(float64)
	roomID := uint(roomIDFloat)
	c.RoomID = roomID
	
	p.stateMu.Lock()
	if _, ok := p.RoomsState[roomID]; !ok {
		gameRoomState := &model.GameRoomState{
			RoomID:           roomID,
			CreatedBy:        c.UserId,
			Players:          []uint{c.UserId},
			Lives:            map[uint]int{c.UserId: DefaultPlayerLives},
			Points:           map[uint]int{c.UserId: DefaultPlayerPoints},
			TurnIndex:        0,
			CharSet:          "",
			Started:          false,
			Round:            0,
			TimeLimit:        0,
			WinnerID:         0,
			CountdownStarted: true,
			CountdownEndTime: time.Now().Add(DefaultCountdownTime),
		}
		p.RoomsState[roomID] = gameRoomState
		p.GameUsecases[roomID] = NewGameUsecase(p, gameRoomState)
		p.GameUsecases[roomID].startRoomCountdown(roomID, gameRoomState)
	} else {
		p.GameUsecases[roomID].addPlayerToRoom(c, p.RoomsState[roomID])
	}
	p.stateMu.Unlock()
	util.RegisterClient(&p.mu, p.Rooms, roomID, c.UserId, c)
	p.GameUsecases[roomID].JoinRoom(c, roomID)
	userJoinedMsg := model.NewUserJoinedMessage(c.UserId, c.UserName, roomID)
	p.GameUsecases[roomID].BroadcastToRoom(roomID, userJoinedMsg)
	return true
}

func (p *GamePool) handleAnswerMessage(c *GameClient, msg model.GameMessage) bool {
	gameRoomState := p.RoomsState[c.RoomID]
	if !p.GameUsecases[c.RoomID].isGameActive(gameRoomState) {
		fmt.Println("Room not found or game not started")
		return false
	}

	answer, ok := p.GameUsecases[c.RoomID].extractStringFromPayload(msg, "answer")
	if !ok {
		fmt.Println("Answer not provided in payload")
		return false
	}

	if p.GameUsecases[c.RoomID].validateAnswer(answer, gameRoomState) {
		p.GameUsecases[c.RoomID].processCorrectAnswer(c, answer, gameRoomState)
	} else {
		p.GameUsecases[c.RoomID].processWrongAnswer(c, answer, gameRoomState)
	}
	return true
}

func (p *GamePool) LeaveRoom(c *GameClient) {
	p.Unregister <- c
}

func (p *GamePool) handleTypingMessage(c *GameClient, msg model.GameMessage) bool {
	gameRoomState := c.Pool.RoomsState[c.RoomID]
	if gameRoomState == nil {
		return false
	}

	text, ok := p.GameUsecases[c.RoomID].extractStringFromPayload(msg, "text")
	if !ok {
		fmt.Println("Text not provided in typing payload")
		return false
	}

	typingMsg := model.NewTypingMessage(text, c.RoomID, c.UserId)
	p.Broadcast <- typingMsg
	return true
}

func (p *GamePool) handlePingMessage(c *GameClient, msg model.GameMessage) {
	if payload, ok := msg.Payload.(map[string]any); ok {
		if timestamp, ok := payload["timestamp"].(float64); ok {
			c.SendPong(int64(timestamp))
		}
	}
}

func (p *GamePool) handlePongMessage(c *GameClient, msg model.GameMessage) {
	if payload, ok := msg.Payload.(map[string]any); ok {
		if timestamp, ok := payload["timestamp"].(float64); ok {
			c.HandlePong(int64(timestamp))
		}
	}
}
