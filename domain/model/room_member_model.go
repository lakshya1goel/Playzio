package model

import "gorm.io/gorm"

type RoomMember struct {
	gorm.Model
	RoomID    uint    `json:"room_id"`
	UserID    *uint   `json:"user_id,omitempty"`
	Username  *string `json:"user_name,omitempty"`
	User      *User   `json:"user,omitempty"`
	GuestID   *string `json:"guest_id,omitempty"`
	GuestName *string `json:"guest_name,omitempty"`
	IsCreator bool    `json:"is_creator"`
}
