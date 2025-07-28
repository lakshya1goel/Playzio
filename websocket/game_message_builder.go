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

func (b *GameMessage) setMessageType(messageType string) *GameMessage {
	b.messageType = messageType
	return b
}

func (b *GameMessage) withTimestamp(timestamp int64) *GameMessage {
	b.payload["timestamp"] = timestamp
	return b
}

func (b *GameMessage) withPingId(pingId int) *GameMessage {
	b.payload["ping_id"] = pingId
	return b
}

func (b *GameMessage) withRoomId(roomId uint) *GameMessage {
	b.payload["room_id"] = roomId
	return b
}

func (b *GameMessage) withUserId(userId uint) *GameMessage {
	b.payload["user_id"] = userId
	return b
}

func (b *GameMessage) withUserName(userName string) *GameMessage {
	b.payload["user_name"] = userName
	return b
}

func (b *GameMessage) withDuration(duration int) *GameMessage {
	b.payload["duration"] = duration
	return b
}

func (b *GameMessage) withCorrect(correct bool) *GameMessage {
	b.payload["correct"] = correct
	return b
}

func (b *GameMessage) withAnswer(answer string) *GameMessage {
	b.payload["answer"] = answer
	return b
}

func (b *GameMessage) withCharSet(charSet string) *GameMessage {
	b.payload["char_set"] = charSet
	return b
}

func (b *GameMessage) withScore(score int) *GameMessage {
	b.payload["score"] = score
	return b
}

func (b *GameMessage) withRound(round int) *GameMessage {
	b.payload["round"] = round
	return b
}

func (b *GameMessage) withLives(lives int) *GameMessage {
	b.payload["lives"] = lives
	return b
}

func (b *GameMessage) withText(text string) *GameMessage {
	b.payload["text"] = text
	return b
}

func (b *GameMessage) withTimeLimit(timeLimit int) *GameMessage {
	b.payload["time_limit"] = timeLimit
	return b
}

func (b *GameMessage) withReason(reason string) *GameMessage {
	b.payload["reason"] = reason
	return b
}

func (b *GameMessage) withWinnerId(winnerId uint) *GameMessage {
	b.payload["winner_id"] = winnerId
	return b
}

func (b *GameMessage) withFinalScores(finalScores map[string]any) *GameMessage {
	b.payload["final_scores"] = finalScores
	return b
}

func (b *GameMessage) build() model.GameMessage {
	return model.GameMessage{
		Type:    b.messageType,
		Payload: b.payload,
	}
}
