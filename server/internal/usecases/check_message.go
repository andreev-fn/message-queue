package usecases

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"server/internal/storage"
	"server/internal/utils"
)

type CheckMsgResult struct {
	ID          string
	Kind        string
	CreatedAt   time.Time
	FinalizedAt *time.Time
	Status      string
	Retries     int
	Payload     json.RawMessage
	Result      *json.RawMessage
}

type CheckMessage struct {
	db              *sql.DB
	msgRepo         *storage.MessageRepository
	archivedMsgRepo *storage.ArchivedMsgRepository
}

func NewCheckMessage(
	db *sql.DB,
	msgRepo *storage.MessageRepository,
	archivedMsgRepo *storage.ArchivedMsgRepository,
) *CheckMessage {
	return &CheckMessage{
		db:              db,
		msgRepo:         msgRepo,
		archivedMsgRepo: archivedMsgRepo,
	}
}

func (uc *CheckMessage) Do(ctx context.Context, id string) (*CheckMsgResult, error) {
	message, err := uc.msgRepo.GetByID(ctx, uc.db, id)
	if err != nil {
		if errors.Is(err, storage.ErrMsgNotFound) {
			return uc.checkArchived(ctx, id)
		}
		return nil, fmt.Errorf("msgRepo.GetByID: %w", err)
	}

	return &CheckMsgResult{
		ID:          message.ID().String(),
		Kind:        message.Kind(),
		Payload:     message.Payload(),
		CreatedAt:   message.CreatedAt(),
		FinalizedAt: message.FinalizedAt(),
		Status:      string(message.Status()),
		Retries:     message.Retries(),
		Result:      message.Result(),
	}, nil
}

func (uc *CheckMessage) checkArchived(ctx context.Context, id string) (*CheckMsgResult, error) {
	archivedMsg, err := uc.archivedMsgRepo.GetByID(ctx, uc.db, id)
	if err != nil {
		return nil, fmt.Errorf("msgRepo.GetByID: %w", err)
	}

	return &CheckMsgResult{
		ID:          archivedMsg.ID().String(),
		Kind:        archivedMsg.Kind(),
		Payload:     archivedMsg.Payload(),
		CreatedAt:   archivedMsg.CreatedAt(),
		FinalizedAt: utils.P(archivedMsg.FinalizedAt()),
		Status:      string(archivedMsg.Status()),
		Retries:     archivedMsg.Retries(),
		Result:      archivedMsg.Result(),
	}, nil
}
