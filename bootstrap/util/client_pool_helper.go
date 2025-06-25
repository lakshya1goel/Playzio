package util

import "sync"

func RegisterClient[T any](mu *sync.RWMutex, poolMap map[uint]map[uint]T, roomID, userID uint, client T) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := poolMap[roomID]; !ok {
		poolMap[roomID] = make(map[uint]T)
	}
	poolMap[roomID][userID] = client
}

func UnregisterClient[T any](mu *sync.RWMutex, poolMap map[uint]map[uint]T, roomID, userID uint) {
	mu.Lock()
	defer mu.Unlock()
	if clients, ok := poolMap[roomID]; ok {
		delete(clients, userID)
		if len(clients) == 0 {
			delete(poolMap, roomID)
		}
	}
}
