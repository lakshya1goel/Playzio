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
			p.mu.Lock()
			if len(p.Rooms[client.RoomID]) >= 10 {
				p.mu.Unlock()
				continue
			}
			util.RegisterClient(&p.mu, p.Rooms, client.RoomID, client.UserId, client)
			if !p.roomSubscriptions[client.RoomID] {
				if err := p.redis.SubscribeToRoom(client.RoomID, func(msg model.ChatMessage) {
					p.Broadcast <- msg
				}); err != nil {
					fmt.Println("Error subscribing to room: ", err)
				}
				p.roomSubscriptions[client.RoomID] = true
			}
			p.mu.Unlock()

		case client := <-p.Unregister:
			p.mu.Lock()
			util.UnregisterClient(&p.mu, p.Rooms, client.RoomID, client.UserId)
			if len(p.Rooms[client.RoomID]) == 0 {
				if err := p.redis.UnsubscribeFromRoom(client.RoomID); err != nil {
					fmt.Println("Error unsubscribing from room: ", err)
				}
				delete(p.roomSubscriptions, client.RoomID)
			}
			p.mu.Unlock()

		case raw := <-p.Broadcast:
			msg, ok := raw.(model.ChatMessage)
			if !ok {
				continue
			}
			msg.CreatedAt = time.Now()
			p.mu.RLock()
			clients := p.Rooms[msg.RoomID]
			for _, client := range clients {
				go client.WriteJSON(msg)
			}
			p.mu.RUnlock()

			go func() {
				if err := p.redis.PublishToRoom(msg.RoomID, msg); err != nil {
					fmt.Println("Error publishing message to room: ", err)
				}
			}()
		}
	}
}
