package requestscope

import (
	"server/internal/domain"
	"server/internal/msgreadiness"
)

type Scope struct {
	Dispatcher       domain.EventDispatcher
	MsgReadyNotifier *msgreadiness.Notifier
}

type Factory interface {
	New() *Scope
}
