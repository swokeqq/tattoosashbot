package conversation

import (
	"context"
	"fmt"

	"github.com/swokeqq/tatusashbot/internal/flow"
)

type ChatKey struct {
	ConnectionID string
	ChatID       int64
}

type StateRepo interface {
	WithState(ctx context.Context, key ChatKey, fn func(flow.State) (flow.State, error)) error
}

type Sender interface {
	Send(ctx context.Context, key ChatKey, action flow.Action) error
}

type Service struct {
	repo   StateRepo
	sender Sender
}

func NewService(repo StateRepo, sender Sender) *Service {
	return &Service{repo: repo, sender: sender}
}

func (s *Service) Handle(ctx context.Context, key ChatKey, event flow.Event) error {
	err := s.repo.WithState(ctx, key, func(current flow.State) (flow.State, error) {
		next, actions := flow.Transition(current, event)

		for _, action := range actions {
			if err := s.sender.Send(ctx, key, action); err != nil {
				return current, fmt.Errorf("send action: %w", err)
			}
		}

		return next, nil
	})
	if err != nil {
		return fmt.Errorf("handle event: %w", err)
	}

	return nil
}
