package eventbus_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"server/internal/eventbus"
	"server/internal/utils/timeutils"
)

type publishedEvent struct {
	Ch  string
	Msg string
}

type mockDriver struct {
	published     []publishedEvent
	handler       eventbus.DriverEventHandler
	listenStarted chan struct{}
	listenerErrCh chan error
}

func newMockDriver() *mockDriver {
	return &mockDriver{
		listenStarted: make(chan struct{}, 1),
		listenerErrCh: make(chan error, 1),
	}
}

func (m *mockDriver) Publish(channel, message string) error {
	m.published = append(m.published, publishedEvent{Ch: channel, Msg: message})
	return nil
}

func (m *mockDriver) Listen(ctx context.Context, _ []string, handler eventbus.DriverEventHandler) error {
	m.handler = handler

	// signal that Listen is ready
	select {
	case m.listenStarted <- struct{}{}:
	default:
	}

	// block until canceled or error triggered
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-m.listenerErrCh:
		return err
	}
}

func (m *mockDriver) Trigger(channel, message string) {
	if m.handler != nil {
		m.handler(channel, message)
	}
}

func (m *mockDriver) TriggerListenerFailure() {
	m.listenerErrCh <- errors.New("test error")
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestEventBus_PublishDelegates(t *testing.T) {
	md := newMockDriver()
	clock := timeutils.NewStubClock(time.Date(2025, 6, 12, 12, 0, 0, 0, time.Local))
	eb := eventbus.NewEventBus(newTestLogger(), clock, md)

	if err := eb.Publish("ch-1", "hello"); err != nil {
		t.Fatalf("Publish error: %v", err)
	}

	if len(md.published) != 1 {
		t.Fatalf("expected 1 publish, got %d", len(md.published))
	}
	if md.published[0].Ch != "ch-1" || md.published[0].Msg != "hello" {
		t.Fatalf("unexpected published: %+v", md.published[0])
	}
}

func TestEventBus_Run_DispatchesToSubscribers(t *testing.T) {
	md := newMockDriver()
	clock := timeutils.NewStubClock(time.Date(2025, 6, 12, 12, 0, 0, 0, time.Local))
	eb := eventbus.NewEventBus(newTestLogger(), clock, md)

	got := make(chan string, 2)
	unsub1 := eb.Subscribe(eventbus.ChannelMsgAvailable, func(message string) { got <- "h1:" + message })
	defer unsub1()
	unsub2 := eb.Subscribe(eventbus.ChannelMsgAvailable, func(message string) { got <- "h2:" + message })
	defer unsub2()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = eb.Run(ctx)
	}()

	// wait until Listen has started
	select {
	case <-md.listenStarted:
	case <-time.After(time.Second):
		t.Fatal("Listen did not start in time")
	}

	md.Trigger(eventbus.ChannelMsgAvailable, "hello")

	// expect both handlers to receive it (order not guaranteed)
	want := map[string]bool{"h1:hello": false, "h2:hello": false}
	for i := 0; i < 2; i++ {
		select {
		case m := <-got:
			if _, ok := want[m]; !ok {
				t.Fatalf("unexpected message: %q", m)
			}
			want[m] = true
		case <-time.After(time.Second):
			t.Fatalf("did not receive handler message %d/2 in time", i+1)
		}
	}

	// Ensure no extra messages arrive
	select {
	case m := <-got:
		t.Fatalf("unexpected extra message: %q", m)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestEventBus_UnsubscribeStopsDelivery(t *testing.T) {
	md := newMockDriver()
	clock := timeutils.NewStubClock(time.Date(2025, 6, 12, 12, 0, 0, 0, time.Local))
	eb := eventbus.NewEventBus(newTestLogger(), clock, md)

	got := make(chan string, 2)

	unsub1 := eb.Subscribe(eventbus.ChannelMsgAvailable, func(message string) { got <- "kept:" + message })
	defer unsub1()
	unsub2 := eb.Subscribe(eventbus.ChannelMsgAvailable, func(message string) { got <- "removed:" + message })
	unsub2() // unsubscribe right away

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = eb.Run(ctx)
	}()

	select {
	case <-md.listenStarted:
	case <-time.After(time.Second):
		t.Fatal("Listen did not start in time")
	}

	md.Trigger(eventbus.ChannelMsgAvailable, "X")

	// Expect only the kept handler to fire
	select {
	case m := <-got:
		if m != "kept:X" {
			t.Fatalf("unexpected message: %q", m)
		}
	case <-time.After(time.Second):
		t.Fatal("no handler message received")
	}

	// Ensure no extra messages arrive
	select {
	case m := <-got:
		t.Fatalf("unexpected extra message: %q", m)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestEventBus_Run_StopsOnContextCancel(t *testing.T) {
	md := newMockDriver()
	clock := timeutils.NewStubClock(time.Date(2025, 6, 12, 12, 0, 0, 0, time.Local))
	eb := eventbus.NewEventBus(newTestLogger(), clock, md)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		_ = eb.Run(ctx)
		close(done)
	}()

	select {
	case <-md.listenStarted:
	case <-time.After(time.Second):
		t.Fatal("Listen did not start in time")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not stop after cancel")
	}
}

func TestEventBus_Run_RestartsOnError(t *testing.T) {
	md := newMockDriver()
	clock := timeutils.NewStubClock(time.Date(2025, 6, 12, 12, 0, 0, 0, time.Local))
	eb := eventbus.NewEventBus(newTestLogger(), clock, md)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = eb.Run(ctx)
	}()

	select {
	case <-md.listenStarted:
	case <-time.After(time.Second):
		t.Fatal("Listen did not start in time")
	}

	clock.Set(clock.Now().Add(time.Minute * 2))

	md.TriggerListenerFailure()

	select {
	case <-md.listenStarted:
	case <-time.After(time.Second):
		t.Fatal("Listen did not restart in time")
	}
}
