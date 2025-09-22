package usecases

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"server/internal/appbuilder/requestscope"
	"server/internal/config"
	"server/internal/domain"
	"server/internal/storage"
	"server/internal/utils/dbutils"
	"server/internal/utils/timeutils"
)

type NewMessageParams struct {
	Queue    string
	Payload  string
	Priority int
	StartAt  *time.Time
}

type PublishMessages struct {
	logger       *slog.Logger
	clock        timeutils.Clock
	db           *sql.DB
	msgRepo      *storage.MessageRepository
	scopeFactory requestscope.Factory
	conf         *config.Config
}

func NewPublishMessages(
	logger *slog.Logger,
	clock timeutils.Clock,
	db *sql.DB,
	msgRepo *storage.MessageRepository,
	scopeFactory requestscope.Factory,
	conf *config.Config,
) *PublishMessages {
	return &PublishMessages{
		logger:       logger,
		clock:        clock,
		db:           db,
		msgRepo:      msgRepo,
		scopeFactory: scopeFactory,
		conf:         conf,
	}
}

func (uc *PublishMessages) Do(
	ctx context.Context,
	messages []NewMessageParams,
	autoRelease bool,
) ([]string, error) {
	if len(messages) > uc.conf.BatchSizeLimit() {
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
		if _, exist := uc.conf.GetQueueConfig(msg.Queue); !exist {
			return nil, fmt.Errorf("queue %s not defined", msg.Queue)
		}

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

	if err := scope.MsgAvailabilityNotifier.Flush(); err != nil {
		uc.logger.Error("scope.MsgAvailabilityNotifier.Flush", "error", err)
	}

	return result, nil
}
