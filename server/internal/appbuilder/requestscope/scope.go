package requestscope

import (
	"server/internal/domain"
	"server/internal/taskreadiness"
)

type Scope struct {
	Dispatcher        domain.EventDispatcher
	TaskReadyNotifier *taskreadiness.Notifier
}

type Factory interface {
	New() *Scope
}
