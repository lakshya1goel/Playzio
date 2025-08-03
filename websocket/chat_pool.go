package websocket

import (
	"fmt"
	"sync"
	"time"

	"github.com/lakshya1goel/Playzio/bootstrap/redis"
	"github.com/lakshya1goel/Playzio/bootstrap/util"
	"github.com/lakshya1goel/Playzio/domain/model"
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
	roomSubscriptions map[uint]bool
	redis             *redis.Redis
}

func NewChatPool(redisClient *redis.Redis) *ChatPool {
	return &ChatPool{
		BasePool:          NewBasePool[*ChatClient](),
		roomSubscriptions: make(map[uint]bool),
		redis:             redisClient,
	}
}

func (p *ChatPool) Start() {
	for {
		select {
		case client := <-p.Register:
			p.handleClientRegister(client)

		case client := <-p.Unregister:
			p.handleClientUnregister(client)

		case raw := <-p.Broadcast:
			if !p.handleBroadcast(raw) {
				continue
			}
		}
	}
}

func (p *ChatPool) handleClientRegister(c *ChatClient) {
	util.RegisterClient(&p.mu, p.Rooms, c.RoomID, c.UserId, c)
	if !p.roomSubscriptions[c.RoomID] {
		if err := p.redis.SubscribeToRoom(c.RoomID, func(msg model.ChatMessage) {
			p.Broadcast <- msg
		}); err != nil {
			fmt.Println("Error subscribing to room: ", err)
		}
		p.roomSubscriptions[c.RoomID] = true
	}
}

func (p *ChatPool) handleClientUnregister(c *ChatClient) {
	util.UnregisterClient(&p.mu, p.Rooms, c.RoomID, c.UserId)
	if len(p.Rooms[c.RoomID]) == 0 {
		if err := p.redis.UnsubscribeFromRoom(c.RoomID); err != nil {
			fmt.Println("Error unsubscribing from room: ", err)
		}
		delete(p.roomSubscriptions, c.RoomID)
	}
}
func (p *ChatPool) handleBroadcast(raw interface{}) bool {
	msg, ok := raw.(model.ChatMessage)
	if !ok {
		return false
	}
	msg.CreatedAt = time.Now()
	p.mu.RLock()
	clients := p.Rooms[msg.RoomID]
	for _, client := range clients {
		go client.WriteJSON(msg)
	}
	p.mu.RUnlock()
	return true
}
