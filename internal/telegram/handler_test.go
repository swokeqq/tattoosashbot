package telegram

import (
	"testing"

	"github.com/go-telegram/bot/models"
	"github.com/swokeqq/tatusashbot/internal/flow"
)

func TestMessageEvent(t *testing.T) {
	const ownerID int64 = 100

	owner := &models.User{ID: ownerID}
	client := &models.User{ID: 200}
	businessBot := &models.User{ID: 300, IsBot: true}

	tests := []struct {
		name      string
		msg       *models.Message
		wantEvent flow.Event
		wantOK    bool
	}{
		{"сообщение клиента", &models.Message{From: client}, flow.Event{Type: flow.EventClientMessage}, true},
		{"сообщение владельца", &models.Message{From: owner}, flow.Event{Type: flow.EventOwnerMessage}, true},
		{"эхо сообщения бота", &models.Message{From: owner, SenderBusinessBot: businessBot}, flow.Event{}, false},
		{"сообщение без отправителя", &models.Message{}, flow.Event{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEvent, gotOK := messageEvent(tt.msg, ownerID)

			if gotOK != tt.wantOK {
				t.Errorf("ok = %v, want %v", gotOK, tt.wantOK)
			}

			if gotEvent != tt.wantEvent {
				t.Errorf("event = %v, want %v", gotEvent, tt.wantEvent)
			}
		})
	}
}
