package appbuilder

import (
	"server/internal/appbuilder/requestscope"
	"server/internal/domain"
	"server/internal/eventbus"
	"server/internal/msgreadiness"
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
	msgReadyNotifier := msgreadiness.NewNotifier(f.eventBus)

	dispatcher := NewEventDispatcher()
	dispatcher.Register(domain.MsgReadyEventName, msgReadyNotifier.HandleEvent)

	return &requestscope.Scope{
		Dispatcher:       dispatcher,
		MsgReadyNotifier: msgReadyNotifier,
	}
}
