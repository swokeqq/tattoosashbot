package conversation

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/swokeqq/tatusashbot/internal/flow"
)

var testKey = ChatKey{ConnectionID: "connection", ChatID: 1}
var testTTL = 14 * 24 * time.Hour

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

func TestHandleFullDialog(t *testing.T) {
	repo := NewMemoryRepo()
	sender := &fakeSender{}
	service := NewService(repo, sender, testTTL)

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
	service := NewService(repo, sender, testTTL)

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
	service := NewService(repo, sender, testTTL)

	setConversation(t, repo, Conversation{State: flow.StateAwaitCustomBrief})

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

func TestHandleReturningClient(t *testing.T) {
	start := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	message := flow.Event{Type: flow.EventClientMessage}
	owner := flow.Event{Type: flow.EventOwnerMessage}
	button := flow.Event{Type: flow.EventButton, Button: flow.ButtonBook}

	tests := []struct {
		name      string
		silence   time.Duration
		event     flow.Event
		wantState flow.State
		wantSent  int
	}{
		{"пауза меньше срока", testTTL - time.Second, message, flow.StateHandedOff, 0},
		{"пауза равна сроку", testTTL, message, flow.StateGreeted, 1},
		{"пауза больше срока", testTTL + time.Hour, message, flow.StateGreeted, 1},
		{"владелец пишет после паузы", testTTL + time.Hour, owner, flow.StateHandedOff, 0},
		{"старая кнопка после паузы", testTTL + time.Hour, button, flow.StateHandedOff, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMemoryRepo()
			sender := &fakeSender{}
			service := NewService(repo, sender, testTTL)
			service.now = func() time.Time { return start.Add(tt.silence) }

			setConversation(t, repo, Conversation{State: flow.StateHandedOff, LastActivity: start})

			if err := service.Handle(context.Background(), testKey, tt.event); err != nil {
				t.Fatalf("handle event: %v", err)
			}

			if got := currentState(t, repo); got != tt.wantState {
				t.Errorf("state = %q, want %q", got, tt.wantState)
			}
			if got := len(sender.sent); got != tt.wantSent {
				t.Errorf("sent = %d, want %d", got, tt.wantSent)
			}
		})
	}
}

func TestHandleActivityExtendsHandoff(t *testing.T) {
	start := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	repo := NewMemoryRepo()
	sender := &fakeSender{}
	service := NewService(repo, sender, testTTL)

	setConversation(t, repo, Conversation{State: flow.StateHandedOff, LastActivity: start})

	service.now = func() time.Time { return start.Add(10 * 24 * time.Hour) }
	if err := service.Handle(context.Background(), testKey, flow.Event{Type: flow.EventOwnerMessage}); err != nil {
		t.Fatalf("handle owner message: %v", err)
	}

	service.now = func() time.Time { return start.Add(20 * 24 * time.Hour) }
	if err := service.Handle(context.Background(), testKey, flow.Event{Type: flow.EventClientMessage}); err != nil {
		t.Fatalf("handle client message: %v", err)
	}

	if got := currentState(t, repo); got != flow.StateHandedOff {
		t.Errorf("state = %q, want %q", got, flow.StateHandedOff)
	}
	if got := len(sender.sent); got != 0 {
		t.Errorf("sent = %d, want 0", got)
	}
}

func currentState(t *testing.T, repo StateRepo) flow.State {
	t.Helper()

	var state flow.State
	err := repo.WithState(context.Background(), testKey, func(current Conversation) (Conversation, error) {
		state = current.State
		return current, nil
	})
	if err != nil {
		t.Fatalf("read state: %v", err)
	}

	return state
}

func setConversation(t *testing.T, repo StateRepo, conversation Conversation) {
	t.Helper()

	err := repo.WithState(context.Background(), testKey, func(Conversation) (Conversation, error) {
		return conversation, nil
	})

	if err != nil {
		t.Fatalf("set conversation: %v", err)
	}
}
