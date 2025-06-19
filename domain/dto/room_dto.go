package dto

import "github.com/lakshya1goel/Playzio/domain/model"

type CreateRoomRequest struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"type" binding:"required"`
}

type CreateRoomResponse struct {
	ID        uint         `json:"id"`
	Name      string       `json:"name"`
	Type      string       `json:"type"`
	CreatedBy uint         `json:"created_by"`
	JoinCode  string       `json:"join_code,omitempty"`
	Members   []model.User `json:"members,omitempty"`
}
