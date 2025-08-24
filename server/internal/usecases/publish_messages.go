package usecases

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"server/internal/appbuilder/requestscope"
	"server/internal/domain"
	"server/internal/storage"
	"server/internal/utils/dbutils"
	"server/internal/utils/timeutils"
)

type NewMessageParams struct {
	Queue    string
	Payload  json.RawMessage
	Priority int
	StartAt  *time.Time
}

type PublishMessages struct {
	logger       *slog.Logger
	clock        timeutils.Clock
	db           *sql.DB
	msgRepo      *storage.MessageRepository
	scopeFactory requestscope.Factory
	maxBatchSize int
}

func NewPublishMessages(
	logger *slog.Logger,
	clock timeutils.Clock,
	db *sql.DB,
	msgRepo *storage.MessageRepository,
	scopeFactory requestscope.Factory,
	maxBatchSize int,
) *PublishMessages {
	return &PublishMessages{
		logger:       logger,
		clock:        clock,
		db:           db,
		msgRepo:      msgRepo,
		scopeFactory: scopeFactory,
		maxBatchSize: maxBatchSize,
	}
}

func (uc *PublishMessages) Do(
	ctx context.Context,
	messages []NewMessageParams,
	autoRelease bool,
) ([]string, error) {
	if len(messages) > uc.maxBatchSize {
		return []string{}, errors.New("batch size limit exceeded")
	}

	scope := uc.scopeFactory.New()

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer dbutils.RollbackWithLog(tx, uc.logger)

	result := make([]string, 0, len(messages))

	for _, msg := range messages {
		message, err := domain.NewMessage(uc.clock, uuid.New(), msg.Queue, msg.Payload, msg.Priority, msg.StartAt)
		if err != nil {
			return nil, err
		}

		if autoRelease {
			if err := message.Release(uc.clock, scope.Dispatcher); err != nil {
				return nil, fmt.Errorf("message.Release: %w", err)
			}
		}

		if err := uc.msgRepo.Save(ctx, tx, message); err != nil {
			return nil, fmt.Errorf("msgRepo.Save: %w", err)
		}

		result = append(result, message.ID().String())
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("tx.Commit: %w", err)
	}

	if err := scope.MsgReadyNotifier.Flush(); err != nil {
		uc.logger.Error("scope.MsgReadyNotifier.Flush", "error", err)
	}

	return result, nil
}
