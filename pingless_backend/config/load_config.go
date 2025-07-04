package config

import (
	"log"
	"strconv"

	"github.com/caarlos0/env/v11"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

type Config struct {
	InviteOnly bool   `env:"INVITE_ONLY"`
	Port       int    `env:"PORT" envDefault:"8080"`
	Email      string `env:"EMAIL" env-required:"true"`
	Password   string `env:"PASSWORD" env-required:"true"`
	EmailHost  string `env:"EMAIL_HOST" env-required:"true"`
	EmailPort  string `env:"EMAIL_PORT" env-required:"true"`
	GifAllowed string `env:"GIF_ALLOWED" envDefault:"true"`
}

func LoadConfig(db *sqlx.DB) Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	// Save all values into DB
	saveSetting(db, "inviteOnly", boolToStr(cfg.InviteOnly))
	saveSetting(db, "port", strconv.Itoa(cfg.Port))
	saveSetting(db, "email", cfg.Email)
	saveSetting(db, "password", cfg.Password)
	saveSetting(db, "emailHost", cfg.EmailHost)
	saveSetting(db, "emailPort", cfg.EmailPort)
	saveSetting(db, "GifAllowed", cfg.GifAllowed)
	return cfg
}

func saveSetting(db *sqlx.DB, key, value string) {
	_, err := db.Exec(`
		INSERT INTO settings (key, value)
		VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)
	if err != nil {
		log.Fatalf("failed to save setting %s: %v", key, err)
	}
}

func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
