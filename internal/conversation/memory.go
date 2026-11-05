package conversation

import (
	"context"
	"fmt"
	"sync"

	"github.com/swokeqq/tatusashbot/internal/flow"
)

type MemoryRepo struct {
	mu     sync.Mutex
	states map[ChatKey]flow.State
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{states: make(map[ChatKey]flow.State)}
}

func (r *MemoryRepo) WithState(_ context.Context, key ChatKey, fn func(flow.State) (flow.State, error)) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.states[key]
	if !ok {
		current = flow.StateNew
	}

	state, err := fn(current)
	if err != nil {
		return fmt.Errorf("transition func: %w", err)
	}

	r.states[key] = state
	return nil
}
