package websocket

import "github.com/lakshya1goel/Playzio/domain/model"

type GameMessage struct {
	messageType string
	payload     map[string]any
}

func NewGameMessage() *GameMessage {
	return &GameMessage{
		payload: make(map[string]any),
	}
}

func (b *GameMessage) SetMessageType(messageType string) *GameMessage {
	b.messageType = messageType
	return b
}

func (b *GameMessage) WithTimestamp(timestamp int64) *GameMessage {
	b.payload["timestamp"] = timestamp
	return b
}

func (b *GameMessage) WithPingId(pingId int) *GameMessage {
	b.payload["ping_id"] = pingId
	return b
}

func (b *GameMessage) WithRoomId(roomId uint) *GameMessage {
	b.payload["room_id"] = roomId
	return b
}

func (b *GameMessage) WithUserId(userId uint) *GameMessage {
	b.payload["user_id"] = userId
	return b
}

func (b *GameMessage) WithUserName(userName string) *GameMessage {
	b.payload["user_name"] = userName
	return b
}

func (b *GameMessage) WithDuration(duration int) *GameMessage {
	b.payload["duration"] = duration
	return b
}

func (b *GameMessage) WithCorrect(correct bool) *GameMessage {
	b.payload["correct"] = correct
	return b
}

func (b *GameMessage) WithAnswer(answer string) *GameMessage {
	b.payload["answer"] = answer
	return b
}

func (b *GameMessage) WithCharSet(charSet string) *GameMessage {
	b.payload["char_set"] = charSet
	return b
}

func (b *GameMessage) WithScore(score int) *GameMessage {
	b.payload["score"] = score
	return b
}

func (b *GameMessage) WithRound(round int) *GameMessage {
	b.payload["round"] = round
	return b
}

func (b *GameMessage) WithLives(lives int) *GameMessage {
	b.payload["lives"] = lives
	return b
}

func (b *GameMessage) WithText(text string) *GameMessage {
	b.payload["text"] = text
	return b
}

func (b *GameMessage) WithTimeLimit(timeLimit int) *GameMessage {
	b.payload["time_limit"] = timeLimit
	return b
}

func (b *GameMessage) WithReason(reason string) *GameMessage {
	b.payload["reason"] = reason
	return b
}

func (b *GameMessage) WithWinnerId(winnerId uint) *GameMessage {
	b.payload["winner_id"] = winnerId
	return b
}

func (b *GameMessage) WithFinalScores(finalScores map[string]any) *GameMessage {
	b.payload["final_scores"] = finalScores
	return b
}

func (b *GameMessage) Build() model.GameMessage {
	return model.GameMessage{
		Type:    b.messageType,
		Payload: b.payload,
	}
}
