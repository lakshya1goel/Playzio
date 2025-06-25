package bootstrap

import "github.com/lakshya1goel/Playzio/domain/model"

type Application struct {
	Env      *Env
	ChatPool *model.ChatPool
	GamePool *model.GamePool
}

func App() Application {
	app := &Application{}
	app.Env = NewEnv()
	app.ChatPool = model.NewChatPool()
	app.GamePool = model.NewGamePool()
	go app.ChatPool.Start()
	go app.GamePool.Start()
	return *app
}
