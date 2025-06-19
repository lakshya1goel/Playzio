package model

import (
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	Type    string `json:"type"`
	Body    string `json:"body"`
	Sender  uint   `json:"sender"`
	RoomID uint   `json:"room_id"`
}

const (
	JoinGroup    = "join-group"
	LeaveGroup   = "leave-group"
	ChatMessage  = "message"
)
