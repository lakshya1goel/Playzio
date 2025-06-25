package model

import (
	"sync"
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
			p.mu.Lock()

			if _, exists := p.Clients[client.RoomID]; !exists {
				p.Clients[client.RoomID] = make(map[uint]*GameClient)
			}
			p.Clients[client.RoomID][client.UserId] = client

			if _, exists := p.Rooms[client.RoomID]; !exists {
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
				if _, ok := room.Lives[client.UserId]; !ok {
					room.Players = append(room.Players, client.UserId)
					room.Lives[client.UserId] = 3
					room.Points[client.UserId] = 0
				}
			}

			p.mu.Unlock()

		case client := <-p.Unregister:
			p.mu.Lock()
			if clients, ok := p.Clients[client.RoomID]; ok {
				delete(clients, client.UserId)
				if len(clients) == 0 {
					delete(p.Clients, client.RoomID)
					delete(p.Rooms, client.RoomID)
				}
			}
			p.mu.Unlock()

		case msg := <-p.Broadcast:
			p.mu.RLock()
			clients := p.Clients[msg.RoomID]
			for _, client := range clients {
				go func(c *GameClient) {
					c.mu.Lock()
					defer c.mu.Unlock()
					c.Conn.WriteJSON(msg)
				}(client)
			}
			p.mu.RUnlock()
		}
	}
}
