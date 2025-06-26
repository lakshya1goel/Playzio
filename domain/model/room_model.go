package model

import "gorm.io/gorm"

type Room struct {
	gorm.Model
	Name           string       `json:"name"`
	Type           string       `json:"type"`
	CreatedBy      *uint        `json:"created_by,omitempty"`
	JoinCode       string       `json:"join_code"`
	CreatorGuestID *string      `json:"creator_guest_id,omitempty"`
	Members        []RoomMember `gorm:"foreignKey:RoomID" json:"members,omitempty"`
}
