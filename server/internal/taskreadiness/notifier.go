package taskreadiness

import (
	"server/internal/domain"
	"server/internal/eventbus"
)

type Notifier struct {
	eventBus   *eventbus.EventBus
	readyKinds map[string]struct{}
}

func NewNotifier(eventBus *eventbus.EventBus) *Notifier {
	return &Notifier{
		eventBus:   eventBus,
		readyKinds: make(map[string]struct{}),
	}
}

func (h *Notifier) HandleEvent(event domain.Event) {
	ev, ok := event.(domain.TaskReadyEvent)
	if !ok {
		return
	}

	h.readyKinds[ev.Kind()] = struct{}{}
}

func (h *Notifier) Flush() error {
	for kind := range h.readyKinds {
		if err := h.eventBus.Publish(eventbus.ChannelTaskReady, kind); err != nil {
			return err
		}
	}
	return nil
}
