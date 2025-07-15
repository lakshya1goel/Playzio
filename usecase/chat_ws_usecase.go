package usecase

import (
	"fmt"

	"github.com/lakshya1goel/Playzio/domain/model"
	"github.com/lakshya1goel/Playzio/websocket"
)

type ChatWSUsecase interface {
	JoinRoom(c *websocket.ChatClient, roomID uint)
	LeaveRoom(c *websocket.ChatClient)
	BroadcastMessage(c *websocket.ChatClient, msg model.ChatMessage)
	Read(c *websocket.ChatClient)
}

type chatWSUsecase struct{}

func NewChatWSUsecase() ChatWSUsecase {
	return &chatWSUsecase{}
}

func (u *chatWSUsecase) JoinRoom(c *websocket.ChatClient, roomID uint) {
	c.RoomID = roomID
	c.Pool.Register <- c
}

func (u *chatWSUsecase) LeaveRoom(c *websocket.ChatClient) {
	c.Pool.Unregister <- c
}

func (u *chatWSUsecase) BroadcastMessage(c *websocket.ChatClient, msg model.ChatMessage) {
	msg.Sender = c.UserId
	msg.RoomID = c.RoomID
	c.Pool.Broadcast <- msg
}

func (u *chatWSUsecase) Read(c *websocket.ChatClient) {
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
