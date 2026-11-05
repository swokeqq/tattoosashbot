package conversation

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/swokeqq/tatusashbot/internal/flow"
)

var testKey = ChatKey{ConnectionID: "connection", ChatID: 1}

type fakeSender struct {
	mu   sync.Mutex
	sent []flow.Action
	err  error
}

func (f *fakeSender) Send(_ context.Context, _ ChatKey, action flow.Action) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.err != nil {
		return f.err
	}
	f.sent = append(f.sent, action)

	return nil
}

func currentState(t *testing.T, repo StateRepo) flow.State {
	t.Helper()

	var state flow.State
	err := repo.WithState(context.Background(), testKey, func(current flow.State) (flow.State, error) {
		state = current
		return current, nil
	})
	if err != nil {
		t.Fatalf("read state: %v", err)
	}

	return state
}

func TestHandleFullDialog(t *testing.T) {
	repo := NewMemoryRepo()
	sender := &fakeSender{}
	service := NewService(repo, sender)

	message := flow.Event{Type: flow.EventClientMessage}
	press := func(button string) flow.Event {
		return flow.Event{Type: flow.EventButton, Button: button}
	}

	events := []flow.Event{
		message,
		press(flow.ButtonBook),
		press(flow.ButtonCustom),
		message,
		message,
		press(flow.ButtonReady),
		message,
	}

	for _, event := range events {
		if err := service.Handle(context.Background(), testKey, event); err != nil {
			t.Fatalf("handle event: %v", err)
		}
	}

	if got := currentState(t, repo); got != flow.StateHandedOff {
		t.Errorf("state = %q, want %q", got, flow.StateHandedOff)
	}
	if got := len(sender.sent); got != 5 {
		t.Errorf("sent = %d, want 5", got)
	}
}

func TestHandleSendErrorKeepsState(t *testing.T) {
	repo := NewMemoryRepo()
	sender := &fakeSender{}
	service := NewService(repo, sender)

	boom := errors.New("boom")
	sender.err = boom

	message := flow.Event{Type: flow.EventClientMessage}

	if err := service.Handle(context.Background(), testKey, message); !errors.Is(err, boom) {
		t.Errorf("error %v, want %v", err, boom)
	}

	if got := currentState(t, repo); got != flow.StateNew {
		t.Errorf("state %q, want %q", got, flow.StateNew)
	}
}

func TestHandleConcurrentMessages(t *testing.T) {
	repo := NewMemoryRepo()
	sender := &fakeSender{}
	service := NewService(repo, sender)

	err := repo.WithState(context.Background(), testKey, func(flow.State) (flow.State, error) {
		return flow.StateAwaitCustomBrief, nil
	})
	if err != nil {
		t.Fatalf("set state: %v", err)
	}

	message := flow.Event{Type: flow.EventClientMessage}

	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			if err := service.Handle(context.Background(), testKey, message); err != nil {
				t.Errorf("handle event: %v", err)
			}
		})
	}
	wg.Wait()

	if got := len(sender.sent); got != 1 {
		t.Errorf("sent = %d, want 1", got)
	}
	if got := currentState(t, repo); got != flow.StateCollectingCustomBrief {
		t.Errorf("state = %q, want %q", got, flow.StateCollectingCustomBrief)
	}
}
