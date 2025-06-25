package model

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
