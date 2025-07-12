package model

import (
	"sync"
	"time"

	"github.com/lakshya1goel/Playzio/bootstrap/util"
)

type BasePool[T any] struct {
	Register   chan T
	Unregister chan T
	Rooms      map[uint]map[uint]T
	Broadcast  chan any
	mu         sync.RWMutex
}

func NewBasePool[T any]() *BasePool[T] {
	return &BasePool[T]{
		Register:   make(chan T),
		Unregister: make(chan T),
		Rooms:      make(map[uint]map[uint]T),
		Broadcast:  make(chan any),
	}
}

type ChatPool struct {
	*BasePool[*ChatClient]
}

func NewChatPool() *ChatPool {
	return &ChatPool{
		BasePool: NewBasePool[*ChatClient](),
	}
}

func (p *ChatPool) Start() {
	for {
		select {
		case client := <-p.Register:
			if len(p.Rooms[client.RoomID]) >= 10 {
				p.mu.Unlock()
				continue
			}
			util.RegisterClient(&p.mu, p.Rooms, client.RoomID, client.UserId, client)

		case client := <-p.Unregister:
			util.UnregisterClient(&p.mu, p.Rooms, client.RoomID, client.UserId)

		case raw := <-p.Broadcast:
			msg, ok := raw.(ChatMessage)
			if !ok {
				continue
			}
			p.mu.RLock()
			clients := p.Rooms[msg.RoomID]
			for _, client := range clients {
				go client.WriteJSON(msg)
			}
			p.mu.RUnlock()
		}
	}
}

type GamePool struct {
	*BasePool[*GameClient]
	RoomsState map[uint]*GameRoomState
	stateMu    sync.RWMutex
}

func NewGamePool() *GamePool {
	return &GamePool{
		BasePool:   NewBasePool[*GameClient](),
		RoomsState: make(map[uint]*GameRoomState),
	}
}

func (p *GamePool) Start() {
	for {
		select {
		case client := <-p.Register:
			util.RegisterClient(&p.mu, p.Rooms, client.RoomID, client.UserId, client)

			p.stateMu.Lock()
			if _, ok := p.RoomsState[client.RoomID]; !ok {
				gameRoomState := &GameRoomState{
					RoomID:           client.RoomID,
					CreatedBy:        client.UserId,
					Players:          []uint{client.UserId},
					Lives:            map[uint]int{client.UserId: 3},
					Points:           map[uint]int{client.UserId: 0},
					TurnIndex:        0,
					CharSet:          "",
					Started:          false,
					Round:            0,
					TimeLimit:        0,
					WinnerID:         0,
					CountdownStarted: true,
					CountdownEndTime: time.Now().Add(2 * time.Minute),
				}
				p.RoomsState[client.RoomID] = gameRoomState

				gameRoomState.CountdownTimer = time.AfterFunc(2*time.Minute, func() {
					p.handleCountdownEnd(client.RoomID)
				})

				p.BroadcastTimerStarted(client.RoomID, 120)

				p.BroadcastToRoom(client.RoomID, GameMessage{
					Type:   UserJoined,
					RoomID: client.RoomID,
					UserID: client.UserId,
					Payload: map[string]any{
						"user_id":   client.UserId,
						"user_name": client.UserName,
						"message":   "Room created and joined",
					},
				})
			} else {
				gameRoomState := p.RoomsState[client.RoomID]
				if _, exists := gameRoomState.Lives[client.UserId]; !exists {
					gameRoomState.Players = append(gameRoomState.Players, client.UserId)
					gameRoomState.Lives[client.UserId] = 3
					gameRoomState.Points[client.UserId] = 0

					if gameRoomState.CountdownStarted && !gameRoomState.Started {
						remainingTime := int(time.Until(gameRoomState.CountdownEndTime).Seconds())
						if remainingTime > 0 {
							p.BroadcastTimerStarted(client.RoomID, remainingTime)
						}
					}

					p.BroadcastToRoom(client.RoomID, GameMessage{
						Type:   UserJoined,
						RoomID: client.RoomID,
						UserID: client.UserId,
						Payload: map[string]any{
							"user_id":   client.UserId,
							"user_name": client.UserName,
							"message":   "User joined the room",
						},
					})
				}
			}
			p.stateMu.Unlock()

		case client := <-p.Unregister:
			p.BroadcastToRoom(client.RoomID, GameMessage{
				Type:   UserLeft,
				RoomID: client.RoomID,
				UserID: client.UserId,
				Payload: map[string]any{
					"user_id":   client.UserId,
					"user_name": client.UserName,
					"message":   "User left the room",
				},
			})

			util.UnregisterClient(&p.mu, p.Rooms, client.RoomID, client.UserId)
			p.stateMu.Lock()
			if len(p.Rooms[client.RoomID]) == 0 {
				if gameRoomState := p.RoomsState[client.RoomID]; gameRoomState != nil && gameRoomState.CountdownTimer != nil {
					gameRoomState.CountdownTimer.Stop()
				}
				delete(p.RoomsState, client.RoomID)
			}
			p.stateMu.Unlock()

		case raw := <-p.Broadcast:
			msg, ok := raw.(GameMessage)
			if !ok {
				continue
			}
			p.mu.RLock()
			clients := p.Rooms[msg.RoomID]
			for _, client := range clients {
				go client.WriteJSON(msg)
			}
			p.mu.RUnlock()
		}
	}
}

func (p *GamePool) handleCountdownEnd(roomID uint) {
	p.stateMu.Lock()
	defer p.stateMu.Unlock()

	gameRoomState := p.RoomsState[roomID]
	if gameRoomState == nil || gameRoomState.Started {
		return
	}

	gameRoomState.Started = true
	gameRoomState.CharSet = util.GenerateRandomString(2, 5)
	gameRoomState.Round = 1
	gameRoomState.TimeLimit = 19
	gameRoomState.CountdownStarted = false
	gameRoomState.TurnIndex = 0

	p.BroadcastToRoom(roomID, GameMessage{
		CharSet: &gameRoomState.CharSet,
		Type:    StartGame,
		RoomID:  roomID,
		Payload: map[string]any{
			"message":    "Game has started",
			"char_set":   gameRoomState.CharSet,
			"round":      gameRoomState.Round,
			"time_limit": gameRoomState.TimeLimit,
		},
	})

	p.BroadcastToRoom(roomID, GameMessage{
		Type:   NextTurn,
		RoomID: roomID,
		Payload: map[string]any{
			"auto_start": true,
		},
	})
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

func (p *GamePool) BroadcastTimerStarted(roomID uint, duration int) {
	p.mu.RLock()
	clients := p.Rooms[roomID]
	if len(clients) == 0 {
		p.mu.RUnlock()
		return
	}
	for _, client := range clients {
		go client.WriteJSON(GameMessage{
			Type:   TimerStarted,
			RoomID: roomID,
			Payload: map[string]any{
				"duration": duration,
			},
		})
	}
	p.mu.RUnlock()
}

func (p *BasePool[T]) RoomCount(roomID uint) int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.Rooms[roomID])
}

func (p *GamePool) BroadcastToRoom(roomID uint, msg GameMessage) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, client := range p.Rooms[roomID] {
		go client.WriteJSON(msg)
	}
}
