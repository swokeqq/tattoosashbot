package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/swokeqq/tatusashbot/internal/conversation"
	"github.com/swokeqq/tatusashbot/internal/flow"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func TestWithStateNewChat(t *testing.T) {
	repo := NewConversationRepo(testPool)

	got := readConversation(t, repo, newTestKey(t))

	if got.State != flow.StateNew {
		t.Errorf("state = %q, want %q", got.State, flow.StateNew)
	}
}

func TestWithStateSaves(t *testing.T) {
	repo := NewConversationRepo(testPool)
	key := newTestKey(t)
	lastActivity := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	saveConversation(t, repo, key, conversation.Conversation{State: flow.StateHandedOff, LastActivity: lastActivity})

	got := readConversation(t, repo, key)

	if got.State != flow.StateHandedOff {
		t.Errorf("state = %q, want %q", got.State, flow.StateHandedOff)
	}
	if !got.LastActivity.Equal(lastActivity) {
		t.Errorf("last activity = %v, want %v", got.LastActivity, lastActivity)
	}
}

func TestWithStateErrorRollsBack(t *testing.T) {
	repo := NewConversationRepo(testPool)
	key := newTestKey(t)
	boom := errors.New("boom")

	err := repo.WithState(context.Background(), key, func(conversation.Conversation) (conversation.Conversation, error) {
		return conversation.Conversation{State: flow.StateHandedOff}, boom
	})
	if !errors.Is(err, boom) {
		t.Errorf("error = %v, want %v", err, boom)
	}

	if got := readConversation(t, repo, key); got.State != flow.StateNew {
		t.Errorf("state = %q, want %q", got.State, flow.StateNew)
	}
}

func TestWithStateIsolatesChats(t *testing.T) {
	repo := NewConversationRepo(testPool)
	first := conversation.ChatKey{ConnectionID: t.Name(), ChatID: 1}
	second := conversation.ChatKey{ConnectionID: t.Name(), ChatID: 2}

	_ = readConversation(t, repo, second)
	saveConversation(t, repo, first, conversation.Conversation{State: flow.StateHandedOff, LastActivity: time.Now()})

	if got := readConversation(t, repo, second); got.State != flow.StateNew {
		t.Errorf("state = %q, want %q", got.State, flow.StateNew)
	}
}

func TestWithStateSerializesChat(t *testing.T) {
	repo := NewConversationRepo(testPool)
	key := newTestKey(t)

	saveConversation(t, repo, key, conversation.Conversation{State: flow.StateAwaitCustomBrief, LastActivity: time.Now()})

	var mu sync.Mutex
	entered := 0

	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			err := repo.WithState(context.Background(), key, func(current conversation.Conversation) (conversation.Conversation, error) {
				if current.State == flow.StateAwaitCustomBrief {
					mu.Lock()
					entered++
					mu.Unlock()
					time.Sleep(50 * time.Millisecond)
				}

				return conversation.Conversation{State: flow.StateCollectingCustomBrief, LastActivity: time.Now()}, nil
			})
			if err != nil {
				t.Errorf("with state: %v", err)
			}
		})
	}
	wg.Wait()

	if entered != 1 {
		t.Errorf("entered = %d, want 1", entered)
	}
}

func runTests(m *testing.M) int {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, "postgres:17-alpine",
		tcpostgres.WithDatabase("test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "start postgres:", err)
		return 1
	}
	defer func() { _ = testcontainers.TerminateContainer(container) }()

	url, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintln(os.Stderr, "connection string:", err)
		return 1
	}

	testPool, err = Connect(ctx, url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect:", err)
		return 1
	}
	defer testPool.Close()

	if err := Migrate(ctx, testPool); err != nil {
		fmt.Fprintln(os.Stderr, "migrate:", err)
		return 1
	}

	return m.Run()
}

func newTestKey(t *testing.T) conversation.ChatKey {
	t.Helper()

	return conversation.ChatKey{ConnectionID: t.Name(), ChatID: 1}
}

func readConversation(t *testing.T, repo *ConversationRepo, key conversation.ChatKey) conversation.Conversation {
	t.Helper()

	var got conversation.Conversation
	err := repo.WithState(context.Background(), key, func(current conversation.Conversation) (conversation.Conversation, error) {
		got = current
		return current, nil
	})
	if err != nil {
		t.Fatalf("read conversation: %v", err)
	}

	return got
}

func saveConversation(t *testing.T, repo *ConversationRepo, key conversation.ChatKey, next conversation.Conversation) {
	t.Helper()

	err := repo.WithState(context.Background(), key, func(conversation.Conversation) (conversation.Conversation, error) {
		return next, nil
	})
	if err != nil {
		t.Fatalf("save conversation: %v", err)
	}
}
