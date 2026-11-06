package config

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	BotToken    string        `env:"BOT_TOKEN,required,notEmpty"`
	OwnerID     int64         `env:"OWNER_USER_ID,required,notEmpty"`
	LogLevel    slog.Level    `env:"LOG_LEVEL" envDefault:"info"`
	HandoffTTL  time.Duration `env:"HANDOFF_TTL" envDefault:"336h"`
	DatabaseURL string        `env:"DATABASE_URL,required,notEmpty"`
}

func Load() (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return cfg, nil
}
