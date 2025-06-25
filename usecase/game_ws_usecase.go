package usecase

import (
	"fmt"

	"github.com/lakshya1goel/Playzio/domain/model"
)

type GameWSUsecase interface {
	JoinRoom(c *model.GameClient, roomID uint)
	LeaveRoom(c *model.GameClient)
	BroadcastMessage(c *model.GameClient, msg model.GameMessage)
	Read(c *model.GameClient)
}

type gameWSUsecase struct{}

func NewGameWSUsecase() GameWSUsecase {
	return &gameWSUsecase{}
}

func (u *gameWSUsecase) JoinRoom(c *model.GameClient, roomID uint) {
	c.RoomID = roomID
	c.Pool.Register <- c
}

func (u *gameWSUsecase) LeaveRoom(c *model.GameClient) {
	c.Pool.Unregister <- c
}

func (u *gameWSUsecase) BroadcastMessage(c *model.GameClient, msg model.GameMessage) {
	msg.UserID = c.UserId
	msg.RoomID = c.RoomID
	c.Pool.Broadcast <- msg
}

func (u *gameWSUsecase) Read(c *model.GameClient) {
	defer func() {
		u.LeaveRoom(c)
		c.Conn.Close()
	}()

	for {
		var msg model.GameMessage
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Game WebSocket read error:", err)
			return
		}

		switch msg.Type {
		case model.Join:
			if msg.RoomID == 0 {
				fmt.Println("Invalid Room ID received")
				continue
			}
			u.JoinRoom(c, msg.RoomID)

		case model.Answer:
			u.BroadcastMessage(c, msg)

		case model.Timeout:
			u.BroadcastMessage(c, msg)

		case model.Leave:
			u.LeaveRoom(c)

		default:
			fmt.Println("Unknown game message type:", msg.Type)
		}
	}
}
