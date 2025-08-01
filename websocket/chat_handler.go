package websocket

import (
	"fmt"
	"time"

	"github.com/lakshya1goel/Playzio/domain/model"
)

type ChatHandler interface {
	JoinRoom(c *ChatClient, roomID uint)
	LeaveRoom(c *ChatClient)
	BroadcastMessage(c *ChatClient, msg model.ChatMessage)
	Read(c *ChatClient)
}

type chatHandler struct{}

func NewChatHandler() ChatHandler {
	return &chatHandler{}
}

func (u *chatHandler) JoinRoom(c *ChatClient, roomID uint) {
	c.RoomID = roomID
	c.Pool.Register <- c
}

func (u *chatHandler) LeaveRoom(c *ChatClient) {
	c.Pool.Unregister <- c
}

func (u *chatHandler) BroadcastMessage(c *ChatClient, msg model.ChatMessage) {
	msg.Sender = c.UserId
	msg.RoomID = c.RoomID
	msg.CreatedAt = time.Now()
	c.Pool.Broadcast <- msg
}

func (u *chatHandler) Read(c *ChatClient) {
	defer func() {
		u.LeaveRoom(c)
		c.Conn.Close()
	}()

	for {
		var msg model.ChatMessage
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Read error:", err)
			return
		}

		switch msg.Type {
		case model.JoinRoom:
			if msg.RoomID == 0 {
				fmt.Println("Error: Invalid Room ID received.")
				continue
			}
			u.JoinRoom(c, msg.RoomID)

		case model.LeaveRoom:
			u.LeaveRoom(c)

		case model.ChatContent:
			if c.RoomID == 0 {
				fmt.Println("Client has not joined any room")
				continue
			}
			u.BroadcastMessage(c, msg)

		default:
			fmt.Println("Unknown message type:", msg.Type)
		}
	}
}
