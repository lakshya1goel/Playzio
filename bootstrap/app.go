package bootstrap

import (
	"fmt"

	"github.com/lakshya1goel/Playzio/bootstrap/redis"
	"github.com/lakshya1goel/Playzio/websocket"
)

type Application struct {
	Env         *Env
	ChatPool    *websocket.ChatPool
	GamePool    *websocket.GamePool
	RedisClient *redis.Redis
}

func App() Application {
	app := &Application{}
	app.Env = NewEnv()

	err := redis.ConnectRedis(
		app.Env.RedisHost,
		app.Env.RedisPort,
		app.Env.RedisPass,
		app.Env.RedisDB,
	)
	if err != nil {
		fmt.Printf("Redis connection failed: %v. Chat will work without Redis.\n", err)
		app.RedisClient = nil
	} else {
		fmt.Println("Redis connected successfully")
		app.RedisClient = redis.RedisClient
	}

	app.ChatPool = websocket.NewChatPool(app.RedisClient)
	app.GamePool = websocket.NewGamePool()
	go app.ChatPool.Start()
	go app.GamePool.Start()
	return *app
}
