package appbuilder

import "server/internal/domain"

var _ domain.EventDispatcher = (*EventDispatcher)(nil)

type EventHandler func(event domain.Event)

type EventDispatcher struct {
	handlers map[string][]EventHandler
}

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		handlers: make(map[string][]EventHandler),
	}
}

func (d *EventDispatcher) Register(eventName string, handler EventHandler) {
	d.handlers[eventName] = append(d.handlers[eventName], handler)
}

func (d *EventDispatcher) Dispatch(event domain.Event) {
	for _, h := range d.handlers[event.Name()] {
		h(event)
	}
}
