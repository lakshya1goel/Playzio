package bootstrap

import (
	"github.com/lakshya1goel/Playzio/websocket"
)

type Application struct {
	Env      *Env
	ChatPool *websocket.ChatPool
	GamePool *websocket.GamePool
}

func App() Application {
	app := &Application{}
	app.Env = NewEnv()
	app.ChatPool = websocket.NewChatPool()
	app.GamePool = websocket.NewGamePool()
	go app.ChatPool.Start()
	go app.GamePool.Start()
	return *app
}
