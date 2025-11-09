package fixtures

import (
	"context"
	"fmt"

	"server/internal/appbuilder"
	"server/internal/domain"
	"server/internal/usecases"
)

func CreateMsg(app *appbuilder.App, queue string, payload string, priority int) string {
	msgIDs, err := app.PublishMessages.Do(
		context.Background(),
		[]usecases.NewMessageParams{{
			Queue:    domain.UnsafeQueueName(queue),
			Payload:  payload,
			Priority: priority,
			StartAt:  nil,
		}},
		false,
	)
	if err != nil {
		panic(err)
	}

	if len(msgIDs) != 1 {
		panic(fmt.Sprint("expected 1 msgID, got ", msgIDs))
	}

	return msgIDs[0]
}

func CreateAvailableMsg(app *appbuilder.App, queue string, payload string, priority int) string {
	msgID := CreateMsg(app, queue, payload, priority)

	err := app.ReleaseMessages.Do(context.Background(), []string{msgID})
	if err != nil {
		panic(err)
	}

	return msgID
}

func CreateProcessingMsg(app *appbuilder.App, queue string, payload string, priority int) string {
	msgID := CreateAvailableMsg(app, queue, payload, priority)

	result, err := app.ConsumeMessages.Do(context.Background(), domain.UnsafeQueueName(queue), 1, 0)
	if err != nil {
		panic(err)
	}

	if len(result) != 1 || result[0].ID != msgID {
		panic("consumed unexpected message")
	}

	return msgID
}

func CreateDelayedMsg(app *appbuilder.App, queue string, payload string) string {
	msgID := CreateProcessingMsg(app, queue, payload, 100)

	err := app.NackMessages.Do(context.Background(), []usecases.NackParams{{ID: msgID, Redeliver: true}})
	if err != nil {
		panic(err)
	}

	return msgID
}

func CreateDeliveredMsg(app *appbuilder.App, queue string, payload string) string {
	msgID := CreateProcessingMsg(app, queue, payload, 100)

	err := app.AckMessages.Do(context.Background(), []usecases.AckParams{{ID: msgID}})
	if err != nil {
		panic(err)
	}

	return msgID
}

func CreateAvailableMsgWithHistory(app *appbuilder.App, historyQueue, queue, payload string) string {
	msgID := CreateProcessingMsg(app, historyQueue, payload, 100)

	err := app.RedirectMessages.Do(context.Background(), []usecases.RedirectParams{
		{ID: msgID, Destination: domain.UnsafeQueueName(queue)},
	})
	if err != nil {
		panic(err)
	}

	return msgID
}

func CreateArchivedMsg(app *appbuilder.App, queue string, payload string) string {
	msgID := CreateDeliveredMsg(app, queue, payload)

	affected, err := app.ArchiveMessages.Do(context.Background(), 1)
	if err != nil {
		panic(err)
	}

	if affected != 1 {
		panic(fmt.Sprint("affected ", affected, " messages (expected 1)"))
	}

	return msgID
}
