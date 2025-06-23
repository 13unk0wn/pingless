package config

import (
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	InviteOnly bool `env:"INVITE_ONLY"`
	Port       int  `env:"PORT" envDefault:"8080"`
}

func LoadConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	return cfg
}
