package domain

const (
	TaskReadyEventName = "TaskReady"
)

type TaskReadyEvent struct {
	kind string
}

func NewTaskReadyEvent(kind string) TaskReadyEvent {
	return TaskReadyEvent{kind}
}

func (e TaskReadyEvent) Name() string { return TaskReadyEventName }
func (e TaskReadyEvent) Kind() string { return e.kind }
