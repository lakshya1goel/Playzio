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
	Type    string `json:"type"`
	RoomID  uint   `json:"room_id"`
	Answer  string `json:"answer"`
	CharSet string `json:"char_set"`
	UserID  uint   `json:"user_id"`
}

const (
	Join    = "join"
	Answer  = "answer"
	Leave   = "leave"
	Timeout = "timeout"
)
