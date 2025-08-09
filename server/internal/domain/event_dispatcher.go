package domain

type Event interface {
	Name() string
}

type EventDispatcher interface {
	Dispatch(ev Event)
}
