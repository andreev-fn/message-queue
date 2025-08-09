package appbuilder

import (
	"server/internal/appbuilder/requestscope"
	"server/internal/domain"
	"server/internal/eventbus"
	"server/internal/taskreadiness"
)

type RequestScopeFactory struct {
	eventBus *eventbus.EventBus
}

func NewRequestScopeFactory(
	eventBus *eventbus.EventBus,
) *RequestScopeFactory {
	return &RequestScopeFactory{
		eventBus: eventBus,
	}
}

func (f *RequestScopeFactory) New() *requestscope.Scope {
	taskReadyNotifier := taskreadiness.NewNotifier(f.eventBus)

	dispatcher := NewEventDispatcher()
	dispatcher.Register(domain.TaskReadyEventName, taskReadyNotifier.HandleEvent)

	return &requestscope.Scope{
		Dispatcher:        dispatcher,
		TaskReadyNotifier: taskReadyNotifier,
	}
}
