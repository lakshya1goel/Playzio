package model

import (
	"gorm.io/gorm"
)

type ChatMessage struct {
	gorm.Model
	Type   string `json:"type"`
	Body   string `json:"body"`
	Sender uint   `json:"sender"`
	RoomID uint   `json:"room_id"`
}

const (
	JoinRoom    = "join-room"
	LeaveRoom   = "leave-room"
	ChatContent = "chat-content"
)

type GameMessage struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload,omitempty"`
}

const (
	Join         = "join"
	Answer       = "answer"
	Leave        = "leave"
	Typing       = "typing"
	TimerStarted = "timer_started"
	StartGame    = "start_game"
	NextTurn     = "next_turn"
	GameOver     = "game_over"
	UserJoined   = "user_joined"
	UserLeft     = "user_left"
	TurnEnded    = "turn_ended"
	Ping         = "ping"
	Pong         = "pong"
)
