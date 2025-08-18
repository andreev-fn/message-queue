package domain

const (
	MsgReadyEventName = "MessageReady"
)

type MsgReadyEvent struct {
	queue string
}

func NewMsgReadyEvent(queue string) MsgReadyEvent {
	return MsgReadyEvent{queue}
}

func (e MsgReadyEvent) Name() string  { return MsgReadyEventName }
func (e MsgReadyEvent) Queue() string { return e.queue }
