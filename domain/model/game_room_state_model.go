package model

import "time"

type GameRoomState struct {
	RoomID     uint
	CreatedBy  uint
	Players    []uint
	Lives      map[uint]int
	Points     map[uint]int
	TurnIndex  int
	CharSet    string
	StartTimer *time.Timer
	Started    bool
	StartChan  chan bool
}

func (g *GameRoomState) StartCountdown(startGameFunc func(), countdownTick func(int)) {
	if g.Started {
		return
	}

	g.StartChan = make(chan bool, 1)
	duration := 60
	g.StartTimer = time.NewTimer(time.Duration(duration) * time.Second)

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		secondsLeft := duration

		for {
			select {
			case <-ticker.C:
				secondsLeft--
				countdownTick(secondsLeft)

			case <-g.StartTimer.C:
				startGameFunc()
				return

			case <-g.StartChan:
				if !g.StartTimer.Stop() {
					<-g.StartTimer.C
				}
				startGameFunc()
				return
			}

			if secondsLeft <= 0 {
				return
			}
		}
	}()
}

func (g *GameRoomState) HandleManualStart() {
	if g.Started {
		return
	}
	g.StartChan <- true
}
