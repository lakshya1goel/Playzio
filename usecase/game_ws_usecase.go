package usecase

import (
	"fmt"

	"github.com/lakshya1goel/Playzio/bootstrap/util"
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
			room := c.Pool.RoomsState[c.RoomID]
			if room == nil || !room.Started {
				fmt.Println("Room not found or game not started")
				continue
			}

			answerRaw, ok := msg.Payload["answer"]
			if !ok {
				fmt.Println("Answer not provided in payload")
				continue
			}
			answer, ok := answerRaw.(string)
			if !ok {
				fmt.Println("Answer is not a string")
				continue
			}

			game := NewGameUsecase(c.Pool, room)

			if util.ContainsSubstring(answer, room.CharSet) && util.IsWordValid(answer) {
				room.CharSet = util.GenerateRandomString(2, 5)
				room.Points[c.UserId]++

				u.BroadcastMessage(c, model.GameMessage{
					Type:   model.Answer,
					RoomID: c.RoomID,
					UserID: c.UserId,
					Payload: map[string]any{
						"correct":    true,
						"answer":     answer,
						"newCharSet": room.CharSet,
						"score":      room.Points[c.UserId],
					},
				})

				game.StartNextTurn()

			} else {
				room.Lives[c.UserId]--
				fmt.Printf("Invalid answer by %d: %s\n", c.UserId, answer)

				u.BroadcastMessage(c, model.GameMessage{
					Type:   model.Answer,
					RoomID: c.RoomID,
					UserID: c.UserId,
					Payload: map[string]any{
						"correct": false,
						"answer":  answer,
						"lives":   room.Lives[c.UserId],
					},
				})

				if game.checkEndCondition() {
					break
				}

				if room.Lives[c.UserId] == 0 && room.Players[room.TurnIndex] == c.UserId {
					game.StartNextTurn()
				}
			}

		case model.Leave:
			u.LeaveRoom(c)

		case model.StartGame:
			if c.RoomID == 0 {
				fmt.Println("Client has not joined any room")
				continue
			}
			room := c.Pool.RoomsState[c.RoomID]
			if room != nil && room.CreatedBy == c.UserId {
				room.Started = true
				room.CharSet = util.GenerateRandomString(2, 5)

				if room.Lives == nil {
					room.Lives = make(map[uint]int)
				}
				for _, uid := range room.Players {
					room.Lives[uid] = 3
				}

				u.BroadcastMessage(c, model.GameMessage{
					Type:   model.StartGame,
					RoomID: c.RoomID,
					Payload: map[string]any{
						"message":  "Game has started",
						"char_set": room.CharSet,
					},
				})

				game := NewGameUsecase(c.Pool, room)
				game.StartNextTurn()
			}

		case model.Typing:
			room := c.Pool.RoomsState[c.RoomID]
			if room == nil {
				break
			}
			c.Pool.Broadcast <- model.GameMessage{
				Type:   model.Typing,
				RoomID: c.RoomID,
				UserID: c.UserId,
				Payload: map[string]any{
					"text": msg.Payload["text"],
				},
			}

		default:
			fmt.Println("Unknown game message type:", msg.Type)
		}
	}
}
