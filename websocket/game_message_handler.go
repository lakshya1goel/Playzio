package websocket

import (
	"fmt"

	"github.com/lakshya1goel/Playzio/bootstrap/util"
	"github.com/lakshya1goel/Playzio/domain/model"
)

type GameMessageHandler interface {
	HandleJoin(client *GameClient, message model.GameMessage) bool
	HandleAnswer(client *GameClient, message model.GameMessage) bool
	HandleTyping(client *GameClient, message model.GameMessage) bool
	ExtractRoomID(message model.GameMessage) (uint, error)
}

type gameMessageHandler struct {
	pool *GamePool
}

func NewGameMessageHandler(pool *GamePool) GameMessageHandler {
	return &gameMessageHandler{
		pool: pool,
	}
}

func (h *gameMessageHandler) HandleJoin(c *GameClient, msg model.GameMessage) bool {
	roomID, err := h.ExtractRoomID(msg)
	if err != nil {
		fmt.Println("Invalid room ID:", err)
		return false
	}

	h.pool.JoinRoom(c, roomID)
	return true
}

func (h *gameMessageHandler) HandleAnswer(c *GameClient, msg model.GameMessage) bool {
	gameRoomState := h.pool.gameStateManager.GetRoomState(c.RoomID)
	if gameRoomState == nil || !gameRoomState.Started {
		fmt.Println("Room not found or game not started")
		return false
	}

	answer, ok := msg.Payload["answer"].(string)
	if !ok {
		fmt.Println("Answer is not a string")
		return false
	}

	game := NewGameEngine(h.pool, gameRoomState)

	if util.ContainsSubstring(answer, gameRoomState.CharSet) && util.IsWordValid(answer) {
		h.handleCorrectAnswer(c, gameRoomState, answer, game)
	} else {
		h.handleWrongAnswer(c, gameRoomState, answer, game)
	}

	return true
}

func (h *gameMessageHandler) HandleTyping(c *GameClient, msg model.GameMessage) bool {
	text, ok := msg.Payload["text"].(string)
	if !ok {
		return false
	}

	message := NewGameMessage().
		SetMessageType(model.Typing).
		WithRoomId(c.RoomID).
		WithUserId(c.UserId).
		WithText(text).
		Build()

	h.pool.BroadcastMessage(c, message)
	return true
}

func (h *gameMessageHandler) ExtractRoomID(msg model.GameMessage) (uint, error) {
	roomIDRaw, exists := msg.Payload["room_id"]
	if !exists {
		return 0, fmt.Errorf("room ID not found in payload")
	}

	switch v := roomIDRaw.(type) {
	case float64:
		return uint(v), nil
	case uint:
		return v, nil
	case int:
		return uint(v), nil
	default:
		return 0, fmt.Errorf("invalid room ID type: %T", v)
	}
}

func (h *gameMessageHandler) handleCorrectAnswer(c *GameClient, state *model.GameRoomState, answer string, game GameEngine) {
	state.CharSet = util.GenerateRandomWord()
	state.Points[c.UserId]++

	message := NewGameMessage().
		SetMessageType(model.Answer).
		WithRoomId(c.RoomID).
		WithUserId(c.UserId).
		WithCorrect(true).
		WithAnswer(answer).
		WithCharSet(state.CharSet).
		WithScore(state.Points[c.UserId]).
		WithLives(state.Lives[c.UserId]).
		Build()

	h.pool.BroadcastMessage(c, message)
	game.handleSuccessfulAnswer(c.UserId, answer, state.CharSet)
}

func (h *gameMessageHandler) handleWrongAnswer(c *GameClient, state *model.GameRoomState, answer string, game GameEngine) {
	message := NewGameMessage().
		SetMessageType(model.Answer).
		WithRoomId(c.RoomID).
		WithUserId(c.UserId).
		WithCorrect(false).
		WithAnswer(answer).
		WithCharSet(state.CharSet).
		WithScore(state.Points[c.UserId]).
		WithLives(state.Lives[c.UserId]).
		Build()

	h.pool.BroadcastMessage(c, message)
	game.handleWrongAnswer(c.UserId, answer)
}
