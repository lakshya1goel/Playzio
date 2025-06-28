package model

import (
	"sync"

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
				room := &GameRoomState{
					RoomID:    client.RoomID,
					CreatedBy: client.UserId,
					Players:   []uint{client.UserId},
					Lives:     map[uint]int{client.UserId: 3},
					Points:    map[uint]int{client.UserId: 0},
					TurnIndex: 0,
					CharSet:   "",
				}
				p.RoomsState[client.RoomID] = room
				p.BroadcastTimerStarted(client.RoomID)
			} else {
				room := p.RoomsState[client.RoomID]
				if _, exists := room.Lives[client.UserId]; !exists {
					room.Players = append(room.Players, client.UserId)
					room.Lives[client.UserId] = 3
					room.Points[client.UserId] = 0
				}
			}
			p.stateMu.Unlock()

		case client := <-p.Unregister:
			util.UnregisterClient(&p.mu, p.Rooms, client.RoomID, client.UserId)
			p.stateMu.Lock()
			if len(p.Rooms[client.RoomID]) == 0 {
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

func (p *GamePool) BroadcastTimerStarted(roomID uint) {
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
				"duration": 60,
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
