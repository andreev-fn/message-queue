package domain

const (
	MsgReadyEventName = "MessageReady"
)

type MsgReadyEvent struct {
	kind string
}

func NewMsgReadyEvent(kind string) MsgReadyEvent {
	return MsgReadyEvent{kind}
}

func (e MsgReadyEvent) Name() string { return MsgReadyEventName }
func (e MsgReadyEvent) Kind() string { return e.kind }
