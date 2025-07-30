package websocket

import "time"

type GameTimerManager interface {
	StartCountdown(roomID uint, duration time.Duration)
	GetRemainingCountdownTime(roomID uint) int
	StopCountdown(roomID uint)
}

type gameTimerManager struct {
	pool *GamePool
}

func NewGameTimerManager(pool *GamePool) GameTimerManager {
	return &gameTimerManager{
		pool: pool,
	}
}

func (g *gameTimerManager) StartCountdown(roomID uint, duration time.Duration) {
	roomState := g.pool.gameStateManager.GetRoomState(roomID)
	if roomState == nil {
		return
	}

	roomState.CountdownTimer = time.AfterFunc(duration, func() {
		g.pool.handleCountdownEnd(roomID)
	})

	g.pool.BroadcastTimerStarted(roomID, int(duration.Seconds()))
}

func (g *gameTimerManager) StopCountdown(roomID uint) {
	roomState := g.pool.gameStateManager.GetRoomState(roomID)
	if roomState != nil && roomState.CountdownTimer != nil {
		roomState.CountdownTimer.Stop()
	}
}

func (g *gameTimerManager) GetRemainingCountdownTime(roomID uint) int {
	roomState := g.pool.gameStateManager.GetRoomState(roomID)
	if roomState == nil || !roomState.CountdownStarted || roomState.Started {
		return 0
	}

	remainingTime := int(time.Until(roomState.CountdownEndTime).Seconds())
	if remainingTime < 0 {
		return 0
	}
	return remainingTime
}
