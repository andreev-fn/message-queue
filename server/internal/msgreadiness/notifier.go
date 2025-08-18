package msgreadiness

import (
	"server/internal/domain"
	"server/internal/eventbus"
)

type Notifier struct {
	eventBus    *eventbus.EventBus
	readyQueues map[string]struct{}
}

func NewNotifier(eventBus *eventbus.EventBus) *Notifier {
	return &Notifier{
		eventBus:    eventBus,
		readyQueues: make(map[string]struct{}),
	}
}

func (h *Notifier) HandleEvent(event domain.Event) {
	ev, ok := event.(domain.MsgReadyEvent)
	if !ok {
		return
	}

	h.readyQueues[ev.Queue()] = struct{}{}
}

func (h *Notifier) Flush() error {
	for queue := range h.readyQueues {
		if err := h.eventBus.Publish(eventbus.ChannelMsgReady, queue); err != nil {
			return err
		}
	}
	return nil
}
