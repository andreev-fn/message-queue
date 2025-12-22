package runkit

import (
	"context"
	"sync"
)

type Multiple []Runnable

func (m Multiple) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errors := make(chan error, len(m))

	var wg sync.WaitGroup

	for _, service := range m {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errors <- service.Run(ctx)
			cancel() // if one service exited - stop all other services
		}()
	}

	wg.Wait()

	close(errors)

	for err := range errors {
		if err != nil {
			return err
		}
	}

	return nil
}
