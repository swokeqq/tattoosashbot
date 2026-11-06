package telegram

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/swokeqq/tatusashbot/internal/conversation"
	"github.com/swokeqq/tatusashbot/internal/flow"
)

func (b *Bot) Send(ctx context.Context, key conversation.ChatKey, action flow.Action) error {
	params := &bot.SendMessageParams{
		BusinessConnectionID: key.ConnectionID,
		ChatID:               key.ChatID,
		Text:                 action.Text,
		ParseMode:            models.ParseModeHTML,
		LinkPreviewOptions:   &models.LinkPreviewOptions{IsDisabled: bot.True()},
	}
	if len(action.Buttons) > 0 {
		params.ReplyMarkup = keyboard(action.Buttons)
	}

	if _, err := b.api.SendMessage(ctx, params); err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func (b *Bot) removeKeyboard(ctx context.Context, msg *models.Message) {
	_, err := b.api.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
		BusinessConnectionID: msg.BusinessConnectionID,
		ChatID:               msg.Chat.ID,
		MessageID:            msg.ID,
		ReplyMarkup:          models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{}},
	})

	if err != nil {
		b.logger.Error("remove keyboard", "chat_id", msg.Chat.ID, "err", err)
	}
}
