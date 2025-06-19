package domain

type ErrorResponse struct {
	Success bool   `json:"success" default:"false"`
	Message string `json:"message"`
}
