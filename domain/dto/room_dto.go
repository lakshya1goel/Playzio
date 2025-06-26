package dto

type CreateRoomRequest struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"type" binding:"required"`
}

type JoinRoomRequest struct {
	JoinCode string `json:"join_code" binding:"required"`
}
