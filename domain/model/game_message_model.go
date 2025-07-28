package model

type GameMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

const (
	Join         = "join"
	Answer       = "answer"
	Leave        = "leave"
	Typing       = "typing"
	TimerStarted = "timer_started"
	StartGame    = "start_game"
	NextTurn     = "next_turn"
	GameOver     = "game_over"
	UserJoined   = "user_joined"
	UserLeft     = "user_left"
	TurnEnded    = "turn_ended"
	Ping         = "ping"
	Pong         = "pong"
)

type AnswerResponsePayload struct {
	Correct    bool   `json:"correct"`
	Answer     string `json:"answer"`
	NewCharSet string `json:"new_char_set,omitempty"`
	Score      int    `json:"score,omitempty"`
	Lives      int    `json:"lives,omitempty"`
	RoomID     uint   `json:"room_id"`
	UserID     uint   `json:"user_id"`
}

type TypingPayload struct {
	Text   string `json:"text"`
	RoomID uint   `json:"room_id"`
	UserID uint   `json:"user_id"`
}

type TimerStartedPayload struct {
	RoomID   uint `json:"room_id"`
	Duration int  `json:"duration"`
}

type StartGamePayload struct {
	RoomID    uint   `json:"room_id"`
	CharSet   string `json:"char_set"`
	Round     int    `json:"round"`
	TimeLimit int    `json:"time_limit"`
}

type UserJoinedPayload struct {
	UserID   uint   `json:"user_id"`
	UserName string `json:"user_name"`
	RoomID   uint   `json:"room_id"`
}

type UserLeftPayload struct {
	UserID   uint   `json:"user_id"`
	UserName string `json:"user_name"`
	RoomID   uint   `json:"room_id"`
}

type PingPayload struct {
	Timestamp int64 `json:"timestamp"`
	PingID    int   `json:"ping_id,omitempty"`
}

type PongPayload struct {
	Timestamp int64 `json:"timestamp"`
}

type GameOverPayload struct {
	RoomID      uint           `json:"room_id"`
	WinnerID    uint           `json:"winner_id"`
	FinalScores map[string]any `json:"final_scores"`
}

type NextTurnPayload struct {
	RoomID    uint   `json:"room_id"`
	UserID    uint   `json:"user_id"`
	CharSet   string `json:"char_set"`
	TimeLimit int    `json:"time_limit"`
	Round     int    `json:"round"`
}

type TurnEndedPayload struct {
	RoomID    uint   `json:"room_id"`
	UserID    uint   `json:"user_id"`
	Reason    string `json:"reason"`
	LivesLeft int    `json:"lives_left"`
	Round     int    `json:"round"`
	Score     int    `json:"score"`
}

func NewAnswerResponseMessage(correct bool, answer string, roomID, userID uint, newCharSet string, score, lives int) GameMessage {
	return GameMessage{
		Type: Answer,
		Payload: AnswerResponsePayload{
			Correct:    correct,
			Answer:     answer,
			NewCharSet: newCharSet,
			Score:      score,
			Lives:      lives,
			RoomID:     roomID,
			UserID:     userID,
		},
	}
}

func NewTypingMessage(text string, roomID, userID uint) GameMessage {
	return GameMessage{
		Type: Typing,
		Payload: TypingPayload{
			Text:   text,
			RoomID: roomID,
			UserID: userID,
		},
	}
}

func NewTimerStartedMessage(roomID uint, duration int) GameMessage {
	return GameMessage{
		Type: TimerStarted,
		Payload: TimerStartedPayload{
			RoomID:   roomID,
			Duration: duration,
		},
	}
}

func NewStartGameMessage(roomID uint, charSet string, round, timeLimit int) GameMessage {
	return GameMessage{
		Type: StartGame,
		Payload: StartGamePayload{
			RoomID:    roomID,
			CharSet:   charSet,
			Round:     round,
			TimeLimit: timeLimit,
		},
	}
}

func NewUserJoinedMessage(userID uint, userName string, roomID uint) GameMessage {
	return GameMessage{
		Type: UserJoined,
		Payload: UserJoinedPayload{
			UserID:   userID,
			UserName: userName,
			RoomID:   roomID,
		},
	}
}

func NewUserLeftMessage(userID uint, userName string, roomID uint) GameMessage {
	return GameMessage{
		Type: UserLeft,
		Payload: UserLeftPayload{
			UserID:   userID,
			UserName: userName,
			RoomID:   roomID,
		},
	}
}

func NewPingMessage(timestamp int64, pingID int) GameMessage {
	return GameMessage{
		Type: Ping,
		Payload: PingPayload{
			Timestamp: timestamp,
			PingID:    pingID,
		},
	}
}

func NewPongMessage(timestamp int64) GameMessage {
	return GameMessage{
		Type: Pong,
		Payload: PongPayload{
			Timestamp: timestamp,
		},
	}
}

func NewGameOverMessage(roomID, winnerID uint, finalScores map[string]any) GameMessage {
	return GameMessage{
		Type: GameOver,
		Payload: GameOverPayload{
			RoomID:      roomID,
			WinnerID:    winnerID,
			FinalScores: finalScores,
		},
	}
}

func NewNextTurnMessage(roomID, userID uint, charSet string, timeLimit, round int) GameMessage {
	return GameMessage{
		Type: NextTurn,
		Payload: NextTurnPayload{
			RoomID:    roomID,
			UserID:    userID,
			CharSet:   charSet,
			TimeLimit: timeLimit,
			Round:     round,
		},
	}
}

func NewTurnEndedMessage(roomID, userID uint, reason string, livesLeft, round, score int) GameMessage {
	return GameMessage{
		Type: TurnEnded,
		Payload: TurnEndedPayload{
			RoomID:    roomID,
			UserID:    userID,
			Reason:    reason,
			LivesLeft: livesLeft,
			Round:     round,
			Score:     score,
		},
	}
}
