package model

import (
	"sync"

	"github.com/lakshya1goel/Playzio/bootstrap/util"
)

type ChatPool struct {
	Register   chan *ChatClient
	Unregister chan *ChatClient
	Rooms      map[uint]map[uint]*ChatClient
	Broadcast  chan ChatMessage
	mu         sync.RWMutex
}

func NewChatPool() *ChatPool {
	return &ChatPool{
		Register:   make(chan *ChatClient),
		Unregister: make(chan *ChatClient),
		Rooms:      make(map[uint]map[uint]*ChatClient),
		Broadcast:  make(chan ChatMessage),
	}
}

func (p *ChatPool) Start() {
	for {
		select {
		case client := <-p.Register:
			util.RegisterClient(&p.mu, p.Rooms, client.RoomID, client.UserId, client)

		case client := <-p.Unregister:
			util.UnregisterClient(&p.mu, p.Rooms, client.RoomID, client.UserId)

		case msg := <-p.Broadcast:
			p.mu.RLock()
			clients := p.Rooms[msg.RoomID]
			for _, client := range clients {
				go client.WriteJSON(msg)
			}
			p.mu.RUnlock()
		}
	}
}
