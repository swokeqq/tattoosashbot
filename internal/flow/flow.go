package flow

type State string

const (
	StateNew                   State = "new"
	StateGreeted               State = "greeted"
	StateChoosingBranch        State = "choosing_branch"
	StateSketchMenu            State = "sketch_menu"
	StateAwaitSketchPhoto      State = "await_sketch_photo"
	StateAwaitTakenSketch      State = "await_taken_sketch"
	StateAwaitCustomBrief      State = "await_custom_brief"
	StateAwaitCoverBrief       State = "await_cover_brief"
	StateCollectingSketchPhoto State = "collecting_sketch_photo"
	StateCollectingTakenSketch State = "collecting_taken_sketch"
	StateCollectingCustomBrief State = "collecting_custom_brief"
	StateCollectingCoverBrief  State = "collecting_cover_brief"
	StateHandedOff             State = "handed_off"
)

type EventType int

const (
	EventClientMessage EventType = iota
	EventButton
	EventOwnerMessage
)

type Event struct {
	Type   EventType
	Button string
}

type Button struct {
	Text string
	Data string
}

type Action struct {
	Text    string
	Buttons []Button
}

const (
	ButtonPersonal     = "personal"
	ButtonBook         = "book"
	ButtonSketchMenu   = "sketch_menu"
	ButtonCustom       = "custom"
	ButtonCover        = "cover"
	ButtonAttachSketch = "attach_sketch"
	ButtonSketchTaken  = "sketch_taken"
	ButtonReady        = "ready"
)

func Transition(s State, e Event) (State, []Action) {
	switch e.Type {
	case EventOwnerMessage:
		return StateHandedOff, nil
	case EventButton:
		return onButton(s, e.Button)
	case EventClientMessage:
		return onClientMessage(s)
	default:
		return s, nil
	}
}

func onButton(s State, button string) (State, []Action) {
	switch s {
	case StateGreeted:
		switch button {
		case ButtonPersonal:
			return StateHandedOff, nil
		case ButtonBook:
			return StateChoosingBranch, reply(textChooseBranch, branchButtons...)
		}
	case StateChoosingBranch:
		switch button {
		case ButtonSketchMenu:
			return StateSketchMenu, reply(textSketchMenu, sketchMenuButtons...)
		case ButtonCustom:
			return StateAwaitCustomBrief, reply(textCustomBrief)
		case ButtonCover:
			return StateAwaitCoverBrief, reply(textCoverBrief)
		}
	case StateSketchMenu:
		switch button {
		case ButtonAttachSketch:
			return StateAwaitSketchPhoto, reply(textAttachSketch)
		case ButtonSketchTaken:
			return StateAwaitTakenSketch, reply(textSketchTaken)
		}
	case StateCollectingSketchPhoto:
		return onReady(s, button, textSketchPhotoDone)
	case StateCollectingTakenSketch:
		return onReady(s, button, textTakenSketchDone)
	case StateCollectingCustomBrief:
		return onReady(s, button, textCustomBriefDone)
	case StateCollectingCoverBrief:
		return onReady(s, button, textCoverBriefDone)
	}

	return s, nil
}

func onClientMessage(s State) (State, []Action) {
	switch s {
	case StateNew:
		return StateGreeted, reply(textGreeting, greetingButtons...)
	case StateGreeted:
		return s, reply(textGreeting, greetingButtons...)
	case StateChoosingBranch:
		return s, reply(textChooseBranch, branchButtons...)
	case StateSketchMenu, StateAwaitSketchPhoto:
		return StateCollectingSketchPhoto, reply(textPressReady, readyButtons...)
	case StateAwaitTakenSketch:
		return StateCollectingTakenSketch, reply(textPressReady, readyButtons...)
	case StateAwaitCustomBrief:
		return StateCollectingCustomBrief, reply(textPressReady, readyButtons...)
	case StateAwaitCoverBrief:
		return StateCollectingCoverBrief, reply(textPressReady, readyButtons...)
	default:
		return s, nil
	}
}

func onReady(s State, button, doneText string) (State, []Action) {
	if button != ButtonReady {
		return s, nil
	}

	return StateHandedOff, reply(doneText)
}

func reply(text string, buttons ...Button) []Action {
	return []Action{{Text: text, Buttons: buttons}}
}
