package msgavailability

import (
	"context"
	"time"

	"server/internal/domain"
)

type Poller struct {
	queue          domain.QueueName
	msgAvailableCh chan struct{}
	pollTimeout    <-chan time.Time
	timedOut       bool
}

func NewPoller(queue domain.QueueName, poll time.Duration) *Poller {
	return &Poller{
		queue:          queue,
		msgAvailableCh: make(chan struct{}, 1),
		pollTimeout:    time.After(poll),
	}
}

func (p *Poller) HandleEvent(message string) {
	if message == p.queue.String() {
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
