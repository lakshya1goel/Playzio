package bootstrap

import "github.com/lakshya1goel/Playzio/domain/model"

type Application struct {
	Env *Env
	Pool *model.Pool
}

func App() Application {
	app := &Application{}
	app.Env = NewEnv()
	app.Pool = model.NewPool()
	go app.Pool.Start()
	return *app
}
