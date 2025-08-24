package msgavailability

import (
	"context"
	"time"
)

type Poller struct {
	queue          string
	msgAvailableCh chan struct{}
	pollTimeout    <-chan time.Time
	timedOut       bool
}

func NewPoller(queue string, poll time.Duration) *Poller {
	return &Poller{
		queue:          queue,
		msgAvailableCh: make(chan struct{}, 1),
		pollTimeout:    time.After(poll),
	}
}

func (p *Poller) HandleEvent(message string) {
	if message == p.queue {
		select {
		case p.msgAvailableCh <- struct{}{}:
		default:
		}
	}
}

func (p *Poller) WaitForNextAttempt(ctx context.Context) {
	select {
	case <-ctx.Done():
		p.timedOut = true
	case <-p.pollTimeout:
		p.timedOut = true
	case <-time.After(time.Second * 3):
	case <-p.msgAvailableCh:
	}
}

func (p *Poller) IsTimedOut() bool {
	return p.timedOut
}
