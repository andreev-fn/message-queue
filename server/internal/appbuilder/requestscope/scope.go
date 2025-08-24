package requestscope

import (
	"server/internal/domain"
	"server/internal/msgavailability"
)

type Scope struct {
	Dispatcher              domain.EventDispatcher
	MsgAvailabilityNotifier *msgavailability.Notifier
}

type Factory interface {
	New() *Scope
}
