package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name       string  `json:"name"`
	Email      string  `json:"email" gorm:"unique"`
	ProfilePic *string `json:"profile_pic"`
	RoomID     *uint   `json:"room_id"`
	Room       *Room    `gorm:"foreignKey:RoomID"`
}
