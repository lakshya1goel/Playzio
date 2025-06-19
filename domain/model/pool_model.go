package model

import (
	"sync"
)

type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Groups     map[uint]map[uint]*Client
	Broadcast  chan Message
	mu         sync.RWMutex
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Groups:     make(map[uint]map[uint]*Client),
		Broadcast:  make(chan Message),
	}
}

func (p *Pool) Start() {
	for {
		select {
		case client := <-p.Register:
			p.mu.Lock()
			if _, exists := p.Groups[client.RoomID]; !exists {
				p.Groups[client.RoomID] = make(map[uint]*Client)
			}
			p.Groups[client.RoomID][client.UserId] = client
			p.mu.Unlock()

		case client := <-p.Unregister:
			p.mu.Lock()
			if clients, ok := p.Groups[client.RoomID]; ok {
				delete(clients, client.UserId)
				if len(clients) == 0 {
					delete(p.Groups, client.RoomID)
				}
			}
			p.mu.Unlock()

		case msg := <-p.Broadcast:
			p.mu.RLock()
			clients := p.Groups[msg.RoomID]
			for _, client := range clients {
				go func(c *Client) {
					c.mu.Lock()
					defer c.mu.Unlock()
					c.Conn.WriteJSON(msg)
				}(client)
			}
			p.mu.RUnlock()
		}
	}
}
