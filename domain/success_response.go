package domain

type SuccessResponse struct {
	Success bool        `json:"success" default:"true"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
