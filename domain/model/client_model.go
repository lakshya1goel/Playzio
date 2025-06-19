package model

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn   *websocket.Conn
	Pool   *Pool
	UserId uint
	RoomID uint
	mu     sync.Mutex
}
