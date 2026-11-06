package postgres

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/swokeqq/tatusashbot/internal/conversation"
	"github.com/swokeqq/tatusashbot/internal/flow"
)

const conversationsTable = "conversations"

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type ConversationRepo struct {
	pool *pgxpool.Pool
}

func NewConversationRepo(pool *pgxpool.Pool) *ConversationRepo {
	return &ConversationRepo{pool: pool}
}

func (r *ConversationRepo) WithState(
	ctx context.Context,
	key conversation.ChatKey,
	fn func(conversation.Conversation) (conversation.Conversation, error)) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	if err := ensureRow(ctx, tx, key); err != nil {
		return err
	}

	current, err := lockRow(ctx, tx, key)
	if err != nil {
		return err
	}

	next, err := fn(current)
	if err != nil {
		return fmt.Errorf("transition func: %w", err)
	}

	if err := updateRow(ctx, tx, key, next); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func ensureRow(ctx context.Context, tx pgx.Tx, key conversation.ChatKey) error {
	query, args, err := psql.Insert(conversationsTable).
		Columns("connection_id", "chat_id", "state").
		Values(key.ConnectionID, key.ChatID, string(flow.StateNew)).
		Suffix("ON CONFLICT (connection_id, chat_id) DO NOTHING").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert: %w", err)
	}

	if _, err := tx.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("insert conversation: %w", err)
	}

	return nil
}

func lockRow(ctx context.Context, tx pgx.Tx, key conversation.ChatKey) (conversation.Conversation, error) {
	query, args, err := psql.Select("state", "updated_at").
		From(conversationsTable).
		Where(sq.Eq{"connection_id": key.ConnectionID, "chat_id": key.ChatID}).
		Suffix("FOR UPDATE").
		ToSql()
	if err != nil {
		return conversation.Conversation{}, fmt.Errorf("build select: %w", err)
	}

	var state string
	var updatedAt time.Time

	if err := tx.QueryRow(ctx, query, args...).Scan(&state, &updatedAt); err != nil {
		return conversation.Conversation{}, fmt.Errorf("select conversation: %w", err)
	}

	return conversation.Conversation{State: flow.State(state), LastActivity: updatedAt}, nil
}

func updateRow(ctx context.Context, tx pgx.Tx, key conversation.ChatKey, next conversation.Conversation) error {
	query, args, err := psql.Update(conversationsTable).
		Set("state", string(next.State)).
		Set("updated_at", next.LastActivity).
		Where(sq.Eq{"connection_id": key.ConnectionID, "chat_id": key.ChatID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("build update: %w", err)
	}

	if _, err := tx.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("update row: %w", err)
	}

	return nil
}
