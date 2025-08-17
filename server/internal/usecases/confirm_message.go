package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"server/internal/appbuilder/requestscope"
	"server/internal/storage"
	"server/internal/utils/timeutils"
)

type ConfirmMessage struct {
	logger       *slog.Logger
	clock        timeutils.Clock
	db           *sql.DB
	msgRepo      *storage.MessageRepository
	scopeFactory requestscope.Factory
}

func NewConfirmMessage(
	logger *slog.Logger,
	clock timeutils.Clock,
	db *sql.DB,
	msgRepo *storage.MessageRepository,
	scopeFactory requestscope.Factory,
) *ConfirmMessage {
	return &ConfirmMessage{
		logger:       logger,
		clock:        clock,
		db:           db,
		msgRepo:      msgRepo,
		scopeFactory: scopeFactory,
	}
}

func (uc *ConfirmMessage) Do(ctx context.Context, id string) error {
	scope := uc.scopeFactory.New()

	message, err := uc.msgRepo.GetByID(ctx, uc.db, id)
	if err != nil {
		return fmt.Errorf("msgRepo.GetByID: %w", err)
	}

	if err := message.Confirm(uc.clock, scope.Dispatcher); err != nil {
		return fmt.Errorf("message.Confirm: %w", err)
	}

	if err := uc.msgRepo.SaveInNewTransaction(ctx, uc.db, message); err != nil {
		return fmt.Errorf("msgRepo.SaveInNewTransaction: %w", err)
	}

	if err := scope.MsgReadyNotifier.Flush(); err != nil {
		uc.logger.Error("scope.MsgReadyNotifier.Flush", "error", err)
	}

	return nil
}
