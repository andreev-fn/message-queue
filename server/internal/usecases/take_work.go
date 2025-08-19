package usecases

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"server/internal/eventbus"
	"server/internal/msgreadiness"
	"server/internal/storage"
	"server/internal/utils/dbutils"
	"server/internal/utils/timeutils"
)

type MessageToWork struct {
	ID      string
	Payload json.RawMessage
}

type TakeWork struct {
	logger   *slog.Logger
	clock    timeutils.Clock
	db       *sql.DB
	msgRepo  *storage.MessageRepository
	eventBus *eventbus.EventBus
}

func NewTakeWork(
	logger *slog.Logger,
	clock timeutils.Clock,
	db *sql.DB,
	msgRepo *storage.MessageRepository,
	eventBus *eventbus.EventBus,
) *TakeWork {
	return &TakeWork{
		logger:   logger,
		clock:    clock,
		db:       db,
		msgRepo:  msgRepo,
		eventBus: eventBus,
	}
}

func (uc *TakeWork) Do(ctx context.Context, queue string, limit int, poll time.Duration) ([]MessageToWork, error) {
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
			return []MessageToWork{}, nil
		}
	}
}

func (uc *TakeWork) takeMessages(ctx context.Context, queue string, limit int) ([]MessageToWork, error) {
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

	var result []MessageToWork

	for _, message := range messages {
		result = append(result, MessageToWork{
			ID:      message.ID().String(),
			Payload: message.Payload(),
		})
	}

	return result, nil
}
