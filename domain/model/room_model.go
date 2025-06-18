package model

import "gorm.io/gorm"

type Room struct {
	gorm.Model
	Name      string `json:"name"`
	Type      string `json:"type"`
	CreatedBy uint   `json:"created_by"`
	Creator   *User   `gorm:"foreignKey:CreatedBy"`
	JoinCode  string `json:"join_code,omitempty"`
	Members   []User `gorm:"foreignKey:RoomID"`
}
