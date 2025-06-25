package model

import (
	"sync"

	"github.com/lakshya1goel/Playzio/bootstrap/util"
)

type GamePool struct {
	Rooms      map[uint]*GameRoomState
	Clients    map[uint]map[uint]*GameClient
	Register   chan *GameClient
	Unregister chan *GameClient
	Broadcast  chan GameMessage
	mu         sync.RWMutex
}

func NewGamePool() *GamePool {
	return &GamePool{
		Rooms:      make(map[uint]*GameRoomState),
		Clients:    make(map[uint]map[uint]*GameClient),
		Register:   make(chan *GameClient),
		Unregister: make(chan *GameClient),
		Broadcast:  make(chan GameMessage),
	}
}

func (p *GamePool) Start() {
	for {
		select {
		case client := <-p.Register:
			util.RegisterClient(&p.mu, p.Clients, client.RoomID, client.UserId, client)

			p.mu.Lock()
			if _, ok := p.Rooms[client.RoomID]; !ok {
				p.Rooms[client.RoomID] = &GameRoomState{
					RoomID:    client.RoomID,
					Players:   []uint{client.UserId},
					Lives:     map[uint]int{client.UserId: 3},
					Points:    map[uint]int{client.UserId: 0},
					TurnIndex: 0,
					CharSet:   "",
				}
			} else {
				room := p.Rooms[client.RoomID]
				if _, exists := room.Lives[client.UserId]; !exists {
					room.Players = append(room.Players, client.UserId)
					room.Lives[client.UserId] = 3
					room.Points[client.UserId] = 0
				}
			}
			p.mu.Unlock()

		case client := <-p.Unregister:
			util.UnregisterClient(&p.mu, p.Clients, client.RoomID, client.UserId)
			p.mu.Lock()
			if len(p.Clients[client.RoomID]) == 0 {
				delete(p.Rooms, client.RoomID)
			}
			p.mu.Unlock()

		case msg := <-p.Broadcast:
			p.mu.RLock()
			clients := p.Clients[msg.RoomID]
			for _, client := range clients {
				go client.WriteJSON(msg)
			}
			p.mu.RUnlock()
		}
	}
}
