package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

type BaseClient struct {
	Conn     *websocket.Conn
	UserId   uint
	UserName string
	RoomID   uint
	mu       sync.Mutex
}

func (bc *BaseClient) WriteJSON(v any) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return bc.Conn.WriteJSON(v)
}

type ChatClient struct {
	BaseClient
	Pool *ChatPool
}

type GameClient struct {
	BaseClient
	Pool *GamePool
}
