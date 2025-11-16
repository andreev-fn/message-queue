package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"server/internal/appbuilder/requestscope"
	"server/internal/config"
	"server/internal/domain"
	"server/internal/storage"
	"server/internal/utils/dbutils"
	"server/internal/utils/timeutils"
)

type RedirectParams struct {
	ID          string
	Destination domain.QueueName
}

type RedirectMessages struct {
	clock        timeutils.Clock
	logger       *slog.Logger
	db           *sql.DB
	msgRepo      *storage.MessageRepository
	scopeFactory requestscope.Factory
	conf         *config.Config
}

func NewRedirectMessages(
	clock timeutils.Clock,
	logger *slog.Logger,
	db *sql.DB,
	msgRepo *storage.MessageRepository,
	scopeFactory requestscope.Factory,
	conf *config.Config,
) *RedirectMessages {
	return &RedirectMessages{
		clock:        clock,
		logger:       logger,
		db:           db,
		msgRepo:      msgRepo,
		scopeFactory: scopeFactory,
		conf:         conf,
	}
}

func (uc *RedirectMessages) Do(ctx context.Context, redirects []RedirectParams) error {
	if len(redirects) > uc.conf.BatchSizeLimit() {
		return ErrBatchSizeTooBig
	}

	scope := uc.scopeFactory.New()

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer dbutils.RollbackWithLog(tx, uc.logger)

	for _, redirect := range redirects {
		// check that the queue exists
		if _, err := uc.conf.GetQueueConfig(redirect.Destination); err != nil {
			return err
		}

		if redirect.Destination.IsDLQ() {
			return ErrDirectWriteToDLQNotAllowed
		}

		message, err := uc.msgRepo.GetByID(ctx, uc.db, redirect.ID)
		if err != nil {
			return fmt.Errorf("msgRepo.GetByID: %w", err)
		}

		if err := message.Redirect(uc.clock, scope.Dispatcher, redirect.Destination); err != nil {
			return fmt.Errorf("message.Redirect: %w", err)
		}

		if err := uc.msgRepo.Save(ctx, tx, message); err != nil {
			return fmt.Errorf("msgRepo.Save: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	if err := scope.MsgAvailabilityNotifier.Flush(); err != nil {
		uc.logger.Error("scope.MsgAvailabilityNotifier.Flush", "error", err)
	}

	return nil
}
