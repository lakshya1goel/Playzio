package bootstrap

import (
	"log"

	"github.com/spf13/viper"
)

type Env struct {
	DBHost string `mapstructure:"DB_HOST"`
	DBPort string `mapstructure:"DB_PORT"`
	DBUser string `mapstructure:"DB_USER"`
	DBPass string `mapstructure:"DB_PASSWORD"`
	DBName string `mapstructure:"DB_NAME"`
}

func NewEnv() *Env {
	env := Env{}
	viper.SetConfigFile(".env")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Can't find the file .env : ", err)
	}

	if err := viper.Unmarshal(&env); err != nil {
		log.Fatal("Environment can't be loaded: ", err)
	}

	return &env
}