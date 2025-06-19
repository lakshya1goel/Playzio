package usecase

import (
	"fmt"

	"github.com/lakshya1goel/Playzio/domain/model"
)

type WSUsecase interface {
	JoinGroup(c *model.Client, groupID uint)
	LeaveGroup(c *model.Client)
	BroadcastMessage(c *model.Client, msg model.Message)
	Read(c *model.Client)
}

type wsUsecase struct{}

func NewWSUsecase() WSUsecase {
	return &wsUsecase{}
}

func (u *wsUsecase) JoinGroup(c *model.Client, roomID uint) {
	c.RoomID = roomID
	c.Pool.Register <- c
}

func (u *wsUsecase) LeaveGroup(c *model.Client) {
	c.Pool.Unregister <- c
}

func (u *wsUsecase) BroadcastMessage(c *model.Client, msg model.Message) {
	msg.Sender = c.UserId
	msg.RoomID = c.RoomID
	c.Pool.Broadcast <- msg
}

func (u *wsUsecase) Read(c *model.Client) {
	defer func() {
		u.LeaveGroup(c)
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
		case model.JoinGroup:
			if msg.RoomID == 0 {
				fmt.Println("Error: Invalid Room ID received.")
				continue
			}
			u.JoinGroup(c, msg.RoomID)

		case model.LeaveGroup:
			u.LeaveGroup(c)

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
