package telegram

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (b *Bot) handleUpdate(ctx context.Context, api *bot.Bot, update *models.Update) {
	if update.Message != nil {
		b.handleMessage(ctx, api, update.Message)
	}
}

func (b *Bot) handleMessage(ctx context.Context, api *bot.Bot, msg *models.Message) {
	if msg.Text == "" {
		return
	}

	_, err := api.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text:   msg.Text,
	})
	if err != nil {
		b.logger.Error("send message", "chat_id", msg.Chat.ID, "err", err)
	}
}
