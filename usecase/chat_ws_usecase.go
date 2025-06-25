package usecase

import (
	"fmt"

	"github.com/lakshya1goel/Playzio/domain/model"
)

type ChatWSUsecase interface {
	JoinRoom(c *model.ChatClient, roomID uint)
	LeaveRoom(c *model.ChatClient)
	BroadcastMessage(c *model.ChatClient, msg model.Message)
	Read(c *model.ChatClient)
}

type chatWSUsecase struct{}

func NewChatWSUsecase() ChatWSUsecase {
	return &chatWSUsecase{}
}

func (u *chatWSUsecase) JoinRoom(c *model.ChatClient, roomID uint) {
	c.RoomID = roomID
	c.Pool.Register <- c
}

func (u *chatWSUsecase) LeaveRoom(c *model.ChatClient) {
	c.Pool.Unregister <- c
}

func (u *chatWSUsecase) BroadcastMessage(c *model.ChatClient, msg model.Message) {
	msg.Sender = c.UserId
	msg.RoomID = c.RoomID
	c.Pool.Broadcast <- msg
}

func (u *chatWSUsecase) Read(c *model.ChatClient) {
	defer func() {
		u.LeaveRoom(c)
		c.Conn.Close()
	}()

	for {
		var msg model.Message
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

		case model.ChatMessage:
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
