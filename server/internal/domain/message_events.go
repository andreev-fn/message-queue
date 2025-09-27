package domain

const (
	MsgAvailableEventName = "MessageAvailable"
)

type MsgAvailableEvent struct {
	queue QueueName
}

func NewMsgAvailableEvent(queue QueueName) MsgAvailableEvent {
	return MsgAvailableEvent{queue}
}

func (e MsgAvailableEvent) Name() string     { return MsgAvailableEventName }
func (e MsgAvailableEvent) Queue() QueueName { return e.queue }
