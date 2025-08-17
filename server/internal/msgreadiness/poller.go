package msgreadiness

import (
	"context"
	"slices"
	"time"
)

type Poller struct {
	kinds       []string
	msgReadyCh  chan struct{}
	pollTimeout <-chan time.Time
	timedOut    bool
}

func NewPoller(kinds []string, poll time.Duration) *Poller {
	return &Poller{
		kinds:       kinds,
		msgReadyCh:  make(chan struct{}, 1),
		pollTimeout: time.After(poll),
	}
}

func (p *Poller) HandleEvent(message string) {
	if slices.Contains(p.kinds, message) {
		select {
		case p.msgReadyCh <- struct{}{}:
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
	case <-p.msgReadyCh:
	}
}

func (p *Poller) IsTimedOut() bool {
	return p.timedOut
}
