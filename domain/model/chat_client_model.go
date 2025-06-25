package model

import (
	"sync"

	"github.com/gorilla/websocket"
)

type ChatClient struct {
	Conn   *websocket.Conn
	Pool   *ChatPool
	UserId uint
	RoomID uint
	mu     sync.Mutex
}
