package websocket

import (
	"sync"
	"time"

	"github.com/lakshya1goel/Playzio/domain/model"
)

type GameStateManager interface {
	CreateRoomState(roomID, userId uint) *model.GameRoomState
	GetRoomState(roomID uint) *model.GameRoomState
	RemoveRoom(roomID uint)
	AddPlayer(roomID, userID uint) bool
}

type gameStateManager struct {
	roomState map[uint]*model.GameRoomState
	mu        sync.RWMutex
}

func NewGameStateManager() GameStateManager {
	return &gameStateManager{
		roomState: make(map[uint]*model.GameRoomState),
	}
}

func (g *gameStateManager) CreateRoomState(roomID, userId uint) *model.GameRoomState {
	g.mu.Lock()
	defer g.mu.Unlock()

	gameRoomState := &model.GameRoomState{
		RoomID:           roomID,
		CreatedBy:        userId,
		Players:          []uint{userId},
		Lives:            map[uint]int{userId: InitialLives},
		Points:           map[uint]int{userId: InitialPoints},
		TurnIndex:        InitialTurnIndex,
		CharSet:          "",
		Started:          false,
		Round:            0,
		TimeLimit:        0,
		WinnerID:         0,
		CountdownStarted: true,
		CountdownEndTime: time.Now().Add(CountdownDuration),
	}

	g.roomState[roomID] = gameRoomState
	return gameRoomState
}

func (g *gameStateManager) GetRoomState(roomID uint) *model.GameRoomState {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.roomState[roomID]
}

func (g *gameStateManager) RemoveRoom(roomID uint) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.roomState, roomID)
}

func (g *gameStateManager) AddPlayer(roomID, userID uint) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	room := g.roomState[roomID]
	if room == nil {
		return false
	}

	if _, exists := room.Lives[userID]; !exists {
		room.Players = append(room.Players, userID)
		room.Lives[userID] = InitialLives
		room.Points[userID] = InitialPoints
		return true
	}
	return false
}
