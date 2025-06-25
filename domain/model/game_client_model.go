package model

import (
	"sync"

	"github.com/gorilla/websocket"
)

type GameClient struct {
	Conn   *websocket.Conn
	Pool   *GamePool
	UserId uint
	RoomID uint
	mu     sync.Mutex
}
