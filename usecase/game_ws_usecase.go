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
	if c.Pool.RoomCount(roomID) >= 10 {
		fmt.Println("Room is full, cannot join:", roomID)
		return
	}	
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

		case model.Answer, model.Timeout:
			if c.RoomID == 0 {
				fmt.Println("Client has not joined any room")
				continue
			}
			u.BroadcastMessage(c, msg)

		case model.Leave:
			u.LeaveRoom(c)

		case model.StartGame:
			if c.RoomID == 0 {
				fmt.Println("Client has not joined any room")
				continue
			}
			roomState := c.Pool.RoomsState[c.RoomID]
			if roomState != nil && roomState.CreatedBy == c.UserId {
				roomState.Started = true
				u.BroadcastMessage(c, model.GameMessage{
					Type:    model.StartGame,
					RoomID:  c.RoomID,
					Payload: map[string]any{"message": "Game has started"},
				})
				fmt.Println("Game manually started by creator:", c.UserId)
			} else {
				fmt.Println("Unauthorized manual start attempt by:", c.UserId)
			}

		default:
			fmt.Println("Unknown game message type:", msg.Type)
		}
	}
}
