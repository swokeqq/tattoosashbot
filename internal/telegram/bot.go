package telegram

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Bot struct {
	api    *bot.Bot
	logger *slog.Logger
}

func New(token string, logger *slog.Logger) (*Bot, error) {
	b := &Bot{logger: logger}

	api, err := bot.New(
		token,
		bot.WithDefaultHandler(b.handleUpdate),
		bot.WithMiddlewares(b.recoverPanic, b.logUpdate),
		bot.WithErrorsHandler(func(err error) {
			logger.Error("telegram api", "err", err)
		}),
		bot.WithAllowedUpdates(bot.AllowedUpdates{
			models.AllowedUpdateMessage,
			models.AllowedUpdateCallbackQuery,
			models.AllowedUpdateBusinessConnection,
			models.AllowedUpdateBusinessMessage,
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}
	b.api = api

	return b, nil
}

func (b *Bot) Run(ctx context.Context) {
	b.api.Start(ctx)
}
