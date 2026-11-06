package telegram

import (
	"context"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const handlerTimeout = 10 * time.Second

func (b *Bot) withTimeout(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, api *bot.Bot, update *models.Update) {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), handlerTimeout)
		defer cancel()

		next(ctx, api, update)
	}
}

func (b *Bot) recoverPanic(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, api *bot.Bot, update *models.Update) {
		defer func() {
			if r := recover(); r != nil {
				b.logger.Error("handler panic", "update_id", update.ID, "panic", r)
			}
		}()
		next(ctx, api, update)
	}
}

func (b *Bot) logUpdate(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, api *bot.Bot, update *models.Update) {
		b.logger.Debug("update received", "update_id", update.ID, "update_type", updateType(update))
		next(ctx, api, update)
	}
}

func updateType(update *models.Update) string {
	switch {
	case update.Message != nil:
		return "message"
	case update.CallbackQuery != nil:
		return "callback_query"
	case update.BusinessConnection != nil:
		return "business_connection"
	case update.BusinessMessage != nil:
		return "business_message"
	default:
		return "unknown"
	}

}
