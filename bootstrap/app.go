package bootstrap

import "github.com/lakshya1goel/Playzio/domain/model"

type Application struct {
	Env *Env
	ChatPool *model.ChatPool
}

func App() Application {
	app := &Application{}
	app.Env = NewEnv()
	app.ChatPool = model.NewPool()
	go app.ChatPool.Start()
	return *app
}
