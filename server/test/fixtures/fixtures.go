package fixtures

import (
	"context"
	"fmt"
	"slices"

	"server/internal/appbuilder"
	"server/internal/domain"
	"server/internal/usecases"
	"server/test/e2eutils"
)

func CreatePreparedMsg(app *appbuilder.App, optArgs ...Option) string {
	opts := buildOptions(optArgs)
	return publish(app, opts.queue, opts.payload, opts.priority, false)
}

func publish(app *appbuilder.App, queue string, payload string, priority int, release bool) string {
	msgIDs, err := app.PublishMessages.Do(
		context.Background(),
		[]usecases.NewMessageParams{{
			Queue:    domain.UnsafeQueueName(queue),
			Payload:  payload,
			Priority: priority,
			StartAt:  nil,
		}},
		release,
	)
	if err != nil {
		panic(err)
	}

	if len(msgIDs) != 1 {
		panic(fmt.Sprint("expected 1 msgID, got ", msgIDs))
	}

	return msgIDs[0]
}

func CreateAvailableMsg(app *appbuilder.App, optArgs ...Option) string {
	opts := buildOptions(optArgs)

	history := append(slices.Clone(opts.history), opts.queue)
	publishQueue, redirectQueues := history[0], history[1:]

	msgID := publish(app, publishQueue, opts.payload, opts.priority, true)

	prevQueue := publishQueue
	for _, nextQueue := range redirectQueues {
		consumeMessage(app, msgID, prevQueue)

		if domain.UnsafeQueueName(nextQueue).IsDLQ() {
			if e2eutils.GetDLQ(prevQueue) != nextQueue {
				panic(fmt.Sprintf("queue %s is unreachable from queue %s", nextQueue, prevQueue))
			}

			nackPermanent(app, msgID)
		} else {
			redirect(app, msgID, nextQueue)
		}

		prevQueue = nextQueue
	}

	return msgID
}

func redirect(app *appbuilder.App, msgID string, toQueue string) {
	if err := app.RedirectMessages.Do(context.Background(), []usecases.RedirectParams{{
		ID:          msgID,
		Destination: domain.UnsafeQueueName(toQueue),
	}}); err != nil {
		panic(err)
	}
}

func nackPermanent(app *appbuilder.App, msgID string) {
	err := app.NackMessages.Do(context.Background(), []usecases.NackParams{{ID: msgID, Redeliver: false}})
	if err != nil {
		panic(err)
	}
}

func CreateProcessingMsg(app *appbuilder.App, optArgs ...Option) string {
	opts := buildOptions(optArgs)

	msgID := CreateAvailableMsg(app, optArgs...)

	consumeMessage(app, msgID, opts.queue)

	return msgID
}

func consumeMessage(app *appbuilder.App, msgID string, queue string) {
	result, err := app.ConsumeMessages.Do(context.Background(), domain.UnsafeQueueName(queue), 1, 0)
	if err != nil {
		panic(err)
	}

	if len(result) != 1 || result[0].ID != msgID {
		panic("consumed unexpected message")
	}
}

func CreateDelayedMsg(app *appbuilder.App, optArgs ...Option) string {
	msgID := CreateProcessingMsg(app, optArgs...)

	err := app.NackMessages.Do(context.Background(), []usecases.NackParams{{ID: msgID, Redeliver: true}})
	if err != nil {
		panic(err)
	}

	return msgID
}

func CreateDeliveredMsg(app *appbuilder.App, optArgs ...Option) string {
	msgID := CreateProcessingMsg(app, optArgs...)

	err := app.AckMessages.Do(context.Background(), []usecases.AckParams{{ID: msgID}})
	if err != nil {
		panic(err)
	}

	return msgID
}

func CreateArchivedMsg(app *appbuilder.App, optArgs ...Option) string {
	msgID := CreateDeliveredMsg(app, optArgs...)

	affected, err := app.ArchiveMessages.Do(context.Background(), 1)
	if err != nil {
		panic(err)
	}

	if affected != 1 {
		panic(fmt.Sprint("affected ", affected, " messages (expected 1)"))
	}

	return msgID
}
