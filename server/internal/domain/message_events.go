package domain

const (
	MsgAvailableEventName = "MessageAvailable"
)

type MsgAvailableEvent struct {
	queue string
}

func NewMsgAvailableEvent(queue string) MsgAvailableEvent {
	return MsgAvailableEvent{queue}
}

func (e MsgAvailableEvent) Name() string  { return MsgAvailableEventName }
func (e MsgAvailableEvent) Queue() string { return e.queue }
