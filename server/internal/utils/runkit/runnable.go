package runkit

import "context"

type Runnable interface {
	Run(ctx context.Context) error
}
