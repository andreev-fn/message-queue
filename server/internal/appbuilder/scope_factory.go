package appbuilder

import (
	"server/internal/appbuilder/requestscope"
	"server/internal/domain"
	"server/internal/eventbus"
	"server/internal/msgavailability"
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
	msgAvailabilityNotifier := msgavailability.NewNotifier(f.eventBus)

	dispatcher := NewEventDispatcher()
	dispatcher.Register(domain.MsgAvailableEventName, msgAvailabilityNotifier.HandleEvent)

	return &requestscope.Scope{
		Dispatcher:              dispatcher,
		MsgAvailabilityNotifier: msgAvailabilityNotifier,
	}
}
