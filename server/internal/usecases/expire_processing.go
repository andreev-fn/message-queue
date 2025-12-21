package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

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

func (uc *ExpireProcessing) Run(ctx context.Context) error {
	for {
		if err := uc.Do(ctx); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
			continue
		}
	}
}

func (uc *ExpireProcessing) Do(ctx context.Context) error {
	const batchSize = 100

	for {
		affected, err := uc.doBatch(ctx, batchSize)
		if err != nil {
			return err
		}

		if affected < batchSize {
			break
		}
	}

	return nil
}

func (uc *ExpireProcessing) doBatch(ctx context.Context, limit int) (int, error) {
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
		if err := message.Nack(uc.clock, scope.Dispatcher, uc.nackPolicy, true); err != nil {
			return 0, fmt.Errorf("message.Nack: %w", err)
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
