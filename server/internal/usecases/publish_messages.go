package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"server/internal/appbuilder/requestscope"
	"server/internal/config"
	"server/internal/domain"
	"server/internal/storage"
	"server/internal/utils/timeutils"
)

type NewMessageParams struct {
	Queue    domain.QueueName
	Payload  string
	Priority int
	StartAt  *time.Time
}

type NewMessageResult struct {
	ID string
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
) ([]BatchResult[NewMessageResult], error) {
	if len(messages) > uc.conf.BatchSizeLimit() {
		return nil, ErrBatchSizeTooBig
	}

	results := make([]BatchResult[NewMessageResult], 0, len(messages))

	for _, params := range messages {
		results = append(results, mapResultToBatch(uc.doOne(ctx, params, autoRelease)))
	}

	return results, nil
}

func (uc *PublishMessages) doOne(
	ctx context.Context,
	params NewMessageParams,
	autoRelease bool,
) (*NewMessageResult, error) {
	scope := uc.scopeFactory.New()

	// check that the queue exists
	if _, err := uc.conf.GetQueueConfig(params.Queue); err != nil {
		return nil, err
	}

	if params.Queue.IsDLQ() {
		return nil, ErrDirectWriteToDLQNotAllowed
	}

	message, err := domain.NewMessage(
		uc.clock,
		uuid.New(),
		params.Queue,
		params.Payload,
		params.Priority,
		params.StartAt,
	)
	if err != nil {
		return nil, err
	}

	if autoRelease {
		if err := message.Release(uc.clock, scope.Dispatcher); err != nil {
			return nil, fmt.Errorf("message.Release: %w", err)
		}
	}

	if err := uc.msgRepo.SaveInNewTransaction(ctx, uc.db, message); err != nil {
		return nil, fmt.Errorf("msgRepo.Save: %w", err)
	}

	if err := scope.MsgAvailabilityNotifier.Flush(); err != nil {
		uc.logger.Error("scope.MsgAvailabilityNotifier.Flush", "error", err)
	}

	return &NewMessageResult{
		ID: message.ID().String(),
	}, nil
}
