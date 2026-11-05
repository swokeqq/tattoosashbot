package flow

type State string

const (
	StateNew              State = "new"
	StateGreeted          State = "greeted"
	StateChoosingBranch   State = "choosing_branch"
	StateSketchMenu       State = "sketch_menu"
	StateAwaitSketchPhoto State = "await_sketch_photo"
	StateAwaitTakenSketch State = "await_taken_sketch"
	StateAwaitCustomBrief State = "await_custom_brief"
	StateAwaitCoverBrief  State = "await_cover_brief"
	StateHandedOff        State = "handed_off"
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
