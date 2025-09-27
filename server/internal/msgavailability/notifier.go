package msgavailability

import (
	"server/internal/domain"
	"server/internal/eventbus"
)

type Notifier struct {
	eventBus        *eventbus.EventBus
	availableQueues map[domain.QueueName]struct{}
}

func NewNotifier(eventBus *eventbus.EventBus) *Notifier {
	return &Notifier{
		eventBus:        eventBus,
		availableQueues: make(map[domain.QueueName]struct{}),
	}
}

func (h *Notifier) HandleEvent(event domain.Event) {
	ev, ok := event.(domain.MsgAvailableEvent)
	if !ok {
		return
	}

	h.availableQueues[ev.Queue()] = struct{}{}
}

func (h *Notifier) Flush() error {
	for queue := range h.availableQueues {
		if err := h.eventBus.Publish(eventbus.ChannelMsgAvailable, queue.String()); err != nil {
			return err
		}
	}
	return nil
}
