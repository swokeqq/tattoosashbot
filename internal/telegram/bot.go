package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/swokeqq/tatusashbot/internal/conversation"
	"github.com/swokeqq/tatusashbot/internal/flow"
)

type EventHandler interface {
	Handle(ctx context.Context, key conversation.ChatKey, event flow.Event) error
}

type Bot struct {
	api         *bot.Bot
	logger      *slog.Logger
	ownerID     int64
	events      EventHandler
	connections sync.Map
}

func New(token string, logger *slog.Logger, ownerID int64) (*Bot, error) {
	b := &Bot{logger: logger, ownerID: ownerID}

	api, err := bot.New(
		token,
		bot.WithDefaultHandler(b.handleUpdate),
		bot.WithErrorsHandler(func(err error) {
			logger.Error("telegram api", "err", err)
		}),
		bot.WithAllowedUpdates(bot.AllowedUpdates{
			models.AllowedUpdateCallbackQuery,
			models.AllowedUpdateBusinessConnection,
			models.AllowedUpdateBusinessMessage,
		}),
		bot.WithNotAsyncHandlers(),
		bot.WithWorkers(4),
		bot.WithMiddlewares(b.recoverPanic, b.withTimeout, b.logUpdate),
	)

	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}
	b.api = api

	return b, nil
}

func (b *Bot) Run(ctx context.Context, events EventHandler) {
	b.events = events
	b.api.Start(ctx)
}
