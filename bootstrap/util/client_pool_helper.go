package util

import "sync"

func RegisterClient[T any](mu *sync.RWMutex, rooms map[uint]map[uint]T, roomID, userID uint, client T) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := rooms[roomID]; !ok {
		rooms[roomID] = make(map[uint]T)
	}
	rooms[roomID][userID] = client
}

func UnregisterClient[T any](mu *sync.RWMutex, rooms map[uint]map[uint]T, roomID, userID uint) {
	mu.Lock()
	defer mu.Unlock()
	if clients, ok := rooms[roomID]; ok {
		delete(clients, userID)
		if len(clients) == 0 {
			delete(rooms, roomID)
		}
	}
}
