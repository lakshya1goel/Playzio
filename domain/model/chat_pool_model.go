package model

import (
	"sync"
)

type ChatPool struct {
	Register   chan *ChatClient
	Unregister chan *ChatClient
	Rooms     map[uint]map[uint]*ChatClient
	Broadcast  chan Message
	mu         sync.RWMutex
}

func NewPool() *ChatPool {
	return &ChatPool{
		Register:   make(chan *ChatClient),
		Unregister: make(chan *ChatClient),
		Rooms:     make(map[uint]map[uint]*ChatClient),
		Broadcast:  make(chan Message),
	}
}

func (p *ChatPool) Start() {
	for {
		select {
		case client := <-p.Register:
			p.mu.Lock()
			if _, exists := p.Rooms[client.RoomID]; !exists {
				p.Rooms[client.RoomID] = make(map[uint]*ChatClient)
			}
			p.Rooms[client.RoomID][client.UserId] = client
			p.mu.Unlock()

		case client := <-p.Unregister:
			p.mu.Lock()
			if clients, ok := p.Rooms[client.RoomID]; ok {
				delete(clients, client.UserId)
				if len(clients) == 0 {
					delete(p.Rooms, client.RoomID)
				}
			}
			p.mu.Unlock()

		case msg := <-p.Broadcast:
			p.mu.RLock()
			clients := p.Rooms[msg.RoomID]
			for _, client := range clients {
				go func(c *ChatClient) {
					c.mu.Lock()
					defer c.mu.Unlock()
					c.Conn.WriteJSON(msg)
				}(client)
			}
			p.mu.RUnlock()
		}
	}
}
