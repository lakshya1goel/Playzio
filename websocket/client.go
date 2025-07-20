package websocket

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lakshya1goel/Playzio/domain/model"
)

type BaseClient struct {
	Conn         *websocket.Conn
	UserId       uint
	UserName     string
	RoomID       uint
	mu           sync.Mutex
	LastPongTime time.Time
	PingInterval time.Duration
	PongTimeout  time.Duration
	PingTicker   *time.Ticker
	PongTimer    *time.Timer
	IsConnected  bool
	PingCount    int
}

func (bc *BaseClient) WriteJSON(v any) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return bc.Conn.WriteJSON(v)
}

func (bc *BaseClient) SendPing() error {
	bc.PingCount++
	pingMsg := model.NewPingMessage(time.Now().Unix(), bc.PingCount)
	return bc.WriteJSON(pingMsg)
}

func (bc *BaseClient) SendPong(timestamp int64) error {
	pongMsg := model.NewPongMessage(timestamp)
	return bc.WriteJSON(pongMsg)
}

func (bc *BaseClient) StartPingPong() {
	bc.PingInterval = 30 * time.Second
	bc.PongTimeout = 10 * time.Second
	bc.IsConnected = true
	bc.LastPongTime = time.Now()
	bc.PingCount = 0

	bc.PingTicker = time.NewTicker(bc.PingInterval)
	bc.PongTimer = time.NewTimer(bc.PongTimeout)

	go bc.pingLoop()
}

func (bc *BaseClient) pingLoop() {
	for bc.IsConnected {
		select {
		case <-bc.PingTicker.C:
			if err := bc.SendPing(); err != nil {
				bc.IsConnected = false
				return
			}
			if bc.PongTimer != nil {
				bc.PongTimer.Stop()
			}
			bc.PongTimer = time.NewTimer(bc.PongTimeout)

		case <-bc.PongTimer.C:
			bc.PongTimer = time.NewTimer(bc.PongTimeout)
		}
	}
}

func (bc *BaseClient) StopPingPong() {
	bc.IsConnected = false
	if bc.PingTicker != nil {
		bc.PingTicker.Stop()
	}
	if bc.PongTimer != nil {
		bc.PongTimer.Stop()
	}
}

func (bc *BaseClient) HandlePong(timestamp int64) {
	bc.LastPongTime = time.Now()
	if bc.PongTimer != nil {
		bc.PongTimer.Stop()
		bc.PongTimer = time.NewTimer(bc.PongTimeout)
	}
}

type ChatClient struct {
	BaseClient
	Pool *ChatPool
}

type GameClient struct {
	BaseClient
	Pool *GamePool
}
