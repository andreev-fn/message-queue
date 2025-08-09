package eventbus

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"server/internal/utils/timeutils"
	"sync"
	"time"
)

const (
	ChannelTaskReady = "task_ready"
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
	for {
		lastAttempt := eb.clock.Now()

		err := eb.driver.Listen(ctx, []string{ChannelTaskReady}, func(channel, message string) {
			eb.dispatch(channel, message)
		})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			eb.logger.Error("eventbus listener failed (will be restarted)", "error", err)
		}

		reconnectDelay := time.Duration(0)
		if eb.clock.Now().Sub(lastAttempt) < time.Minute {
			reconnectDelay = time.Second * 5
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(reconnectDelay):
		}
	}
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
