package flow

import (
	"slices"
	"testing"
)

func TestTransition(t *testing.T) {
	message := Event{Type: EventClientMessage}
	owner := Event{Type: EventOwnerMessage}
	press := func(button string) Event {
		return Event{Type: EventButton, Button: button}
	}

	tests := []struct {
		name        string
		state       State
		event       Event
		wantState   State
		wantText    string
		wantButtons []string
	}{
		{"первое сообщение показывает меню", StateNew, message, StateGreeted, textGreeting, []string{ButtonPersonal, ButtonBook}},
		{"текст вместо кнопки в меню повторяет вопрос", StateGreeted, message, StateGreeted, textGreeting, []string{ButtonPersonal, ButtonBook}},
		{"личный чат передает диалог", StateGreeted, press(ButtonPersonal), StateHandedOff, "", nil},
		{"запись открывает выбор ветки", StateGreeted, press(ButtonBook), StateChoosingBranch, textChooseBranch, []string{ButtonSketchMenu, ButtonCustom, ButtonCover}},
		{"текст вместо кнопки при выборе ветки повторяет вопрос", StateChoosingBranch, message, StateChoosingBranch, textChooseBranch, []string{ButtonSketchMenu, ButtonCustom, ButtonCover}},

		{"просмотр готовых эскизов", StateChoosingBranch, press(ButtonSketchMenu), StateSketchMenu, textSketchMenu, []string{ButtonAttachSketch, ButtonSketchTaken}},
		{"индивидуальный эскиз", StateChoosingBranch, press(ButtonCustom), StateAwaitCustomBrief, textCustomBrief, nil},
		{"перекрытие работы", StateChoosingBranch, press(ButtonCover), StateAwaitCoverBrief, textCoverBrief, nil},
		{"кнопка прикрепить эскиз", StateSketchMenu, press(ButtonAttachSketch), StateAwaitSketchPhoto, textAttachSketch, nil},
		{"кнопка эскиз занят", StateSketchMenu, press(ButtonSketchTaken), StateAwaitTakenSketch, textSketchTaken, nil},

		{"эскиз прислали без нажатия кнопки", StateSketchMenu, message, StateCollectingSketchPhoto, textPressReady, []string{ButtonReady}},
		{"первое сообщение с эскизом", StateAwaitSketchPhoto, message, StateCollectingSketchPhoto, textPressReady, []string{ButtonReady}},
		{"первое сообщение про занятый эскиз", StateAwaitTakenSketch, message, StateCollectingTakenSketch, textPressReady, []string{ButtonReady}},
		{"первое сообщение с описанием эскиза", StateAwaitCustomBrief, message, StateCollectingCustomBrief, textPressReady, []string{ButtonReady}},
		{"первое сообщение про перекрытие", StateAwaitCoverBrief, message, StateCollectingCoverBrief, textPressReady, []string{ButtonReady}},

		{"следующие сообщения с эскизом без ответа", StateCollectingSketchPhoto, message, StateCollectingSketchPhoto, "", nil},
		{"следующие сообщения про занятый эскиз без ответа", StateCollectingTakenSketch, message, StateCollectingTakenSketch, "", nil},
		{"следующие сообщения с описанием без ответа", StateCollectingCustomBrief, message, StateCollectingCustomBrief, "", nil},
		{"следующие сообщения про перекрытие без ответа", StateCollectingCoverBrief, message, StateCollectingCoverBrief, "", nil},

		{"готово завершает ветку эскиза", StateCollectingSketchPhoto, press(ButtonReady), StateHandedOff, textSketchPhotoDone, nil},
		{"готово завершает ветку занятого эскиза", StateCollectingTakenSketch, press(ButtonReady), StateHandedOff, textTakenSketchDone, nil},
		{"готово завершает индивидуальную ветку", StateCollectingCustomBrief, press(ButtonReady), StateHandedOff, textCustomBriefDone, nil},
		{"готово завершает ветку перекрытия", StateCollectingCoverBrief, press(ButtonReady), StateHandedOff, textCoverBriefDone, nil},

		{"старая кнопка игнорируется", StateAwaitCustomBrief, press(ButtonBook), StateAwaitCustomBrief, "", nil},
		{"готово без сообщений игнорируется", StateAwaitCustomBrief, press(ButtonReady), StateAwaitCustomBrief, "", nil},
		{"старая кнопка при сборе сообщений игнорируется", StateCollectingCustomBrief, press(ButtonBook), StateCollectingCustomBrief, "", nil},
		{"кнопка без начатого диалога игнорируется", StateNew, press(ButtonBook), StateNew, "", nil},
		{"кнопка после передачи игнорируется", StateHandedOff, press(ButtonReady), StateHandedOff, "", nil},
		{"после передачи диалога владельцу бот молчит", StateHandedOff, message, StateHandedOff, "", nil},

		{"владелец написал первым", StateNew, owner, StateHandedOff, "", nil},
		{"сообщение владельца останавливает бота", StateChoosingBranch, owner, StateHandedOff, "", nil},
		{"сообщение владельца прерывает сбор сообщений", StateCollectingCustomBrief, owner, StateHandedOff, "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotState, gotActions := Transition(tt.state, tt.event)

			if gotState != tt.wantState {
				t.Errorf("state = %q, want %q", gotState, tt.wantState)
			}

			var gotText string
			var gotButtons []string
			if len(gotActions) > 0 {
				gotText = gotActions[0].Text
				for _, button := range gotActions[0].Buttons {
					gotButtons = append(gotButtons, button.Data)
				}
			}

			if gotText != tt.wantText {
				t.Errorf("text = %q, want %q", gotText, tt.wantText)
			}
			if !slices.Equal(gotButtons, tt.wantButtons) {
				t.Errorf("buttons = %v, want %v", gotButtons, tt.wantButtons)
			}
		})
	}
}
