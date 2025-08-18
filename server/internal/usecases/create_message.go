package usecases

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"server/internal/appbuilder/requestscope"
	"server/internal/domain"
	"server/internal/storage"
	"server/internal/utils/timeutils"
)

type CreateMessage struct {
	logger       *slog.Logger
	clock        timeutils.Clock
	db           *sql.DB
	msgRepo      *storage.MessageRepository
	scopeFactory requestscope.Factory
}

func NewCreateMessage(
	logger *slog.Logger,
	clock timeutils.Clock,
	db *sql.DB,
	msgRepo *storage.MessageRepository,
	scopeFactory requestscope.Factory,
) *CreateMessage {
	return &CreateMessage{
		logger:       logger,
		clock:        clock,
		db:           db,
		msgRepo:      msgRepo,
		scopeFactory: scopeFactory,
	}
}

func (uc *CreateMessage) Do(
	ctx context.Context,
	queue string,
	payload json.RawMessage,
	priority int,
	autoConfirm bool,
	startAt *time.Time,
) (string, error) {
	scope := uc.scopeFactory.New()

	message, err := domain.NewMessage(uc.clock, uuid.New(), queue, payload, priority, startAt)
	if err != nil {
		return "", err
	}

	if autoConfirm {
		if err := message.Confirm(uc.clock, scope.Dispatcher); err != nil {
			return "", fmt.Errorf("message.Confirm: %w", err)
		}
	}

	if err := uc.msgRepo.SaveInNewTransaction(ctx, uc.db, message); err != nil {
		return "", fmt.Errorf("msgRepo.SaveInNewTransaction: %w", err)
	}

	if err := scope.MsgReadyNotifier.Flush(); err != nil {
		uc.logger.Error("scope.MsgReadyNotifier.Flush", "error", err)
	}

	return message.ID().String(), nil
}
