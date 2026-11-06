package conversation

import (
	"context"
	"fmt"
	"time"

	"github.com/swokeqq/tatusashbot/internal/flow"
)

type ChatKey struct {
	ConnectionID string
	ChatID       int64
}

type Conversation struct {
	State        flow.State
	LastActivity time.Time
}

type StateRepo interface {
	WithState(ctx context.Context, key ChatKey, fn func(Conversation) (Conversation, error)) error
}

type Sender interface {
	Send(ctx context.Context, key ChatKey, action flow.Action) error
}

type Service struct {
	repo       StateRepo
	sender     Sender
	handoffTTL time.Duration
	now        func() time.Time
}

func NewService(repo StateRepo, sender Sender, handoffTTL time.Duration) *Service {
	return &Service{repo: repo, sender: sender, handoffTTL: handoffTTL, now: time.Now}
}

func (s *Service) Handle(ctx context.Context, key ChatKey, event flow.Event) error {
	err := s.repo.WithState(ctx, key, func(current Conversation) (Conversation, error) {
		now := s.now()
		effective := event
		if s.isReturning(current, event, now) {
			effective = flow.Event{Type: flow.EventClientReturned}
		}

		next, actions := flow.Transition(current.State, effective)

		for _, action := range actions {
			if err := s.sender.Send(ctx, key, action); err != nil {
				return current, fmt.Errorf("send action: %w", err)
			}
		}

		return Conversation{State: next, LastActivity: now}, nil
	})
	if err != nil {
		return fmt.Errorf("handle event: %w", err)
	}

	return nil
}

func (s *Service) isReturning(current Conversation, event flow.Event, now time.Time) bool {
	return event.Type == flow.EventClientMessage &&
		current.State == flow.StateHandedOff &&
		now.Sub(current.LastActivity) >= s.handoffTTL
}
