package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/swokeqq/tatusashbot/internal/config"
	"github.com/swokeqq/tatusashbot/internal/telegram"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tgBot, err := telegram.New(cfg.BotToken, logger)
	if err != nil {
		return err
	}

	logger.Info("bot started")
	tgBot.Run(ctx)
	logger.Info("bot stopped")

	return nil
}
