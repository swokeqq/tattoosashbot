package telegram

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/swokeqq/tatusashbot/internal/conversation"
	"github.com/swokeqq/tatusashbot/internal/flow"
)

func (b *Bot) handleUpdate(ctx context.Context, api *bot.Bot, update *models.Update) {
	switch {
	case update.BusinessConnection != nil:
		b.handleConnection(update.BusinessConnection)
	case update.BusinessMessage != nil:
		b.handleBusinessMessage(ctx, update.BusinessMessage)
	case update.CallbackQuery != nil:
		b.handleCallback(ctx, update.CallbackQuery)
	}
}

func (b *Bot) handleConnection(conn *models.BusinessConnection) {
	allowed := conn.User.ID == b.ownerID && conn.IsEnabled
	b.connections.Store(conn.ID, allowed)
	b.logger.Info("business connection updated", "connection id", conn.ID, "allowed", allowed)
}

func (b *Bot) handleBusinessMessage(ctx context.Context, msg *models.Message) {
	event, ok := messageEvent(msg, b.ownerID)
	if !ok {
		return
	}

	b.dispatch(ctx, msg, event)
}

func (b *Bot) handleCallback(ctx context.Context, query *models.CallbackQuery) {
	if _, err := b.api.AnswerCallbackQuery(
		ctx,
		&bot.AnswerCallbackQueryParams{CallbackQueryID: query.ID}); err != nil {
		b.logger.Error("answer callback", "err", err)
	}

	msg := query.Message.Message
	if msg == nil || msg.BusinessConnectionID == "" || query.From.ID == b.ownerID {
		return
	}

	if b.dispatch(ctx, msg, flow.Event{Type: flow.EventButton, Button: query.Data}) {
		b.removeKeyboard(ctx, msg)
	}
}

func (b *Bot) isAllowedConnection(ctx context.Context, id string) (bool, error) {
	if allowed, ok := b.connections.Load(id); ok {
		return allowed.(bool), nil
	}

	conn, err := b.api.GetBusinessConnection(ctx, &bot.GetBusinessConnectionParams{BusinessConnectionID: id})
	if err != nil {
		return false, fmt.Errorf("get business connection: %w", err)
	}
	allowed := conn.User.ID == b.ownerID && conn.IsEnabled
	b.connections.Store(id, allowed)

	return allowed, nil
}

func messageEvent(msg *models.Message, ownerID int64) (flow.Event, bool) {
	if msg.SenderBusinessBot != nil || msg.From == nil {
		return flow.Event{}, false
	}
	if msg.From.ID == ownerID {
		return flow.Event{Type: flow.EventOwnerMessage}, true
	}

	return flow.Event{Type: flow.EventClientMessage}, true
}

func (b *Bot) dispatch(ctx context.Context, msg *models.Message, event flow.Event) bool {
	allowed, err := b.isAllowedConnection(ctx, msg.BusinessConnectionID)
	if err != nil {
		b.logger.Error("check connection", "connection id", msg.BusinessConnectionID, "err", err)
		return false
	}
	if !allowed {
		return false
	}

	key := conversation.ChatKey{ConnectionID: msg.BusinessConnectionID, ChatID: msg.Chat.ID}
	if err := b.events.Handle(ctx, key, event); err != nil {
		b.logger.Error("handle event", "chat id", msg.BusinessConnectionID, "err", err)
		return false
	}

	return true
}
