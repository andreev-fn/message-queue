package usecases

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"server/internal/eventbus"
	"server/internal/msgreadiness"
	"server/internal/storage"
	"server/internal/utils/dbutils"
	"server/internal/utils/timeutils"
)

type MessageToConsume struct {
	ID      string
	Payload string
}

type ConsumeMessages struct {
	logger       *slog.Logger
	clock        timeutils.Clock
	db           *sql.DB
	msgRepo      *storage.MessageRepository
	eventBus     *eventbus.EventBus
	maxBatchSize int
}

func NewConsumeMessages(
	logger *slog.Logger,
	clock timeutils.Clock,
	db *sql.DB,
	msgRepo *storage.MessageRepository,
	eventBus *eventbus.EventBus,
	maxBatchSize int,
) *ConsumeMessages {
	return &ConsumeMessages{
		logger:       logger,
		clock:        clock,
		db:           db,
		msgRepo:      msgRepo,
		eventBus:     eventBus,
		maxBatchSize: maxBatchSize,
	}
}

func (uc *ConsumeMessages) Do(
	ctx context.Context,
	queue string,
	limit int,
	poll time.Duration,
) ([]MessageToConsume, error) {
	if limit > uc.maxBatchSize {
		return []MessageToConsume{}, errors.New("batch size limit exceeded")
	}

	// fast path first
	result, err := uc.takeMessages(ctx, queue, limit)
	if err != nil {
		return nil, err
	}

	if len(result) > 0 || poll == 0 {
		return result, nil
	}

	poller := msgreadiness.NewPoller(queue, poll)
	unsubscribe := uc.eventBus.Subscribe(eventbus.ChannelMsgReady, poller.HandleEvent)
	defer unsubscribe()

	for {
		result, err = uc.takeMessages(ctx, queue, limit)
		if err != nil {
			return nil, err
		}

		if len(result) > 0 {
			return result, nil
		}

		poller.WaitForNextAttempt(ctx)
		if poller.IsTimedOut() {
			return []MessageToConsume{}, nil
		}
	}
}

func (uc *ConsumeMessages) takeMessages(ctx context.Context, queue string, limit int) ([]MessageToConsume, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer dbutils.RollbackWithLog(tx, uc.logger)

	messages, err := uc.msgRepo.GetReadyWithLock(ctx, tx, queue, limit)
	if err != nil {
		return nil, fmt.Errorf("msgRepo.GetReadyWithLock: %w", err)
	}

	for _, message := range messages {
		if err := message.StartProcessing(uc.clock); err != nil {
			return nil, fmt.Errorf("message.StartProcessing: %w", err)
		}

		if err := uc.msgRepo.Save(ctx, tx, message); err != nil {
			return nil, fmt.Errorf("msgRepo.Save: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("tx.Commit: %w", err)
	}

	var result []MessageToConsume

	for _, message := range messages {
		result = append(result, MessageToConsume{
			ID:      message.ID().String(),
			Payload: message.Payload(),
		})
	}

	return result, nil
}
