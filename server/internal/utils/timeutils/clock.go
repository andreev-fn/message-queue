package timeutils

import "time"

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func NewRealClock() *RealClock {
	return &RealClock{}
}

func (*RealClock) Now() time.Time {
	return time.Now()
}

type StubClock struct {
	now time.Time
}

func NewStubClock(now time.Time) *StubClock {
	return &StubClock{now: now}
}

func (c *StubClock) Now() time.Time {
	return c.now
}

func (c *StubClock) Set(now time.Time) {
	c.now = now
}
