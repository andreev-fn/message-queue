package eventbus

import (
	"context"
	"log/slog"
	"math"
	"sync"

	"server/internal/utils/timeutils"
)

const (
	ChannelMsgAvailable = "message_available"
)

type EventHandler func(message string)

type EventBus struct {
	logger   *slog.Logger
	clock    timeutils.Clock
	driver   PubSubDriver
	handlers map[string]map[int64]EventHandler
	nextID   int64
	mu       sync.RWMutex
}

func NewEventBus(logger *slog.Logger, clock timeutils.Clock, driver PubSubDriver) *EventBus {
	return &EventBus{
		logger:   logger,
		clock:    clock,
		driver:   driver,
		handlers: make(map[string]map[int64]EventHandler),
	}
}

func (eb *EventBus) Run(ctx context.Context) error {
	return eb.driver.Listen(ctx, []string{ChannelMsgAvailable}, func(channel, message string) {
		eb.dispatch(channel, message)
	})
}

func (eb *EventBus) Publish(channel, message string) error {
	return eb.driver.Publish(channel, message)
}

func (eb *EventBus) Subscribe(channel string, h EventHandler) (unsubscribe func()) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if _, exist := eb.handlers[channel]; !exist {
		eb.handlers[channel] = make(map[int64]EventHandler)
	}

	sid := eb.nextID
	if eb.nextID == math.MaxInt64 {
		eb.nextID = 0
	} else {
		eb.nextID++
	}

	eb.handlers[channel][sid] = h

	return func() {
		eb.mu.Lock()
		defer eb.mu.Unlock()

		delete(eb.handlers[channel], sid)
	}
}

func (eb *EventBus) dispatch(channel, message string) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if _, exist := eb.handlers[channel]; !exist {
		return
	}

	for _, h := range eb.handlers[channel] {
		go h(message)
	}
}
