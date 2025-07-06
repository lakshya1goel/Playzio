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

	remainingTime := c.Pool.GetRemainingCountdownTime(roomID)
	if remainingTime > 0 {
		c.WriteJSON(model.GameMessage{
			Type:   model.TimerStarted,
			RoomID: roomID,
			Payload: map[string]any{
				"duration": remainingTime,
			},
		})
	}
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
			gameRoomState := c.Pool.RoomsState[c.RoomID]
			if gameRoomState == nil || !gameRoomState.Started {
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

			game := NewGameUsecase(c.Pool, gameRoomState)

			if util.ContainsSubstring(answer, gameRoomState.CharSet) && util.IsWordValid(answer) {
				gameRoomState.CharSet = util.GenerateRandomString(2, 5)
				gameRoomState.Points[c.UserId]++

				u.BroadcastMessage(c, model.GameMessage{
					Type:   model.Answer,
					RoomID: c.RoomID,
					UserID: c.UserId,
					Payload: map[string]any{
						"correct":    true,
						"answer":     answer,
						"newCharSet": gameRoomState.CharSet,
						"score":      gameRoomState.Points[c.UserId],
					},
				})

				game.StartNextTurn()

			} else {
				gameRoomState.Lives[c.UserId]--
				fmt.Printf("Invalid answer by %d: %s\n", c.UserId, answer)

				u.BroadcastMessage(c, model.GameMessage{
					Type:   model.Answer,
					RoomID: c.RoomID,
					UserID: c.UserId,
					Payload: map[string]any{
						"correct": false,
						"answer":  answer,
						"lives":   gameRoomState.Lives[c.UserId],
					},
				})

				if game.checkEndCondition() {
					break
				}

				if gameRoomState.Lives[c.UserId] == 0 && gameRoomState.Players[gameRoomState.TurnIndex] == c.UserId {
					game.StartNextTurn()
				}
			}

		case model.NextTurn:
			if autoStart, ok := msg.Payload["auto_start"].(bool); ok && autoStart {
				gameRoomState := c.Pool.RoomsState[c.RoomID]
				if gameRoomState != nil && gameRoomState.Started {
					game := NewGameUsecase(c.Pool, gameRoomState)
					game.StartNextTurn()
				}
			}

		case model.Leave:
			u.LeaveRoom(c)

		case model.Typing:
			gameRoomState := c.Pool.RoomsState[c.RoomID]
			if gameRoomState == nil {
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
