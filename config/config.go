package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	App    AppConfig
	Logger LoggerConfig
}

type AppConfig struct {
	Env  string
	Port int
}

type LoggerConfig struct {
	Level string
}

func LoadEnv() *Config {
	viper.SetConfigFile(".env")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Error reading .env file:", err)
	}

	config := &Config{
		App: AppConfig{
			Env:  viper.GetString("APP_ENV"),
			Port: viper.GetInt("APP_PORT"),
		},
		Logger: LoggerConfig{
			Level: viper.GetString("LOG_LEVEL"),
		},
	}

	return config
}
