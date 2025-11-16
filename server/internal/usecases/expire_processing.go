package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"server/internal/appbuilder/requestscope"
	"server/internal/domain"
	"server/internal/storage"
	"server/internal/utils/dbutils"
	"server/internal/utils/timeutils"
)

type ExpireProcessing struct {
	clock        timeutils.Clock
	logger       *slog.Logger
	db           *sql.DB
	msgRepo      *storage.MessageRepository
	scopeFactory requestscope.Factory
	nackPolicy   *domain.NackPolicy
}

func NewExpireProcessing(
	clock timeutils.Clock,
	logger *slog.Logger,
	db *sql.DB,
	msgRepo *storage.MessageRepository,
	scopeFactory requestscope.Factory,
	nackPolicy *domain.NackPolicy,
) *ExpireProcessing {
	return &ExpireProcessing{
		clock:        clock,
		logger:       logger,
		db:           db,
		msgRepo:      msgRepo,
		scopeFactory: scopeFactory,
		nackPolicy:   nackPolicy,
	}
}

func (uc *ExpireProcessing) Do(ctx context.Context, limit int) (int, error) {
	messages, err := uc.msgRepo.GetProcessingToExpire(ctx, uc.db, limit)
	if err != nil {
		return 0, fmt.Errorf("msgRepo.GetProcessingToExpire: %w", err)
	}

	scope := uc.scopeFactory.New()

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer dbutils.RollbackWithLog(tx, uc.logger)

	for _, message := range messages {
		action, err := uc.nackPolicy.Decide(message, true)
		if err != nil {
			return 0, err
		}

		switch action.Type {
		case domain.NackActionDelay:
			if err := message.Delay(uc.clock, uc.clock.Now().Add(action.DelayDuration)); err != nil {
				return 0, fmt.Errorf("message.Delay: %w", err)
			}
		case domain.NackActionDrop:
			if err := message.MarkDropped(uc.clock); err != nil {
				return 0, fmt.Errorf("message.MarkDropped: %w", err)
			}
		case domain.NackActionDLQ:
			dlQueue, err := message.Queue().DLQName()
			if err != nil {
				return 0, fmt.Errorf("queue.DLQName: %w", err)
			}

			if err := message.Redirect(uc.clock, scope.Dispatcher, dlQueue); err != nil {
				return 0, fmt.Errorf("message.MarkDropped: %w", err)
			}
		}

		if err := uc.msgRepo.Save(ctx, tx, message); err != nil {
			return 0, fmt.Errorf("msgRepo.Save: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("tx.Commit: %w", err)
	}

	if err := scope.MsgAvailabilityNotifier.Flush(); err != nil {
		uc.logger.Error("scope.MsgAvailabilityNotifier.Flush", "error", err)
	}

	return len(messages), nil
}
