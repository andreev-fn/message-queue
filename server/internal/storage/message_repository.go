package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"server/internal/domain"
	"server/internal/utils/dbutils"
	"server/internal/utils/timeutils"
)

const selectAll = `
	SELECT 
		m.id, m.queue, m.created_at, m.finalized_at, m.status, m.status_changed_at,
		m.delayed_until, m.timeout_at, m.priority, m.retries, m.version,
		p.payload
	FROM messages m
	LEFT JOIN message_payloads p ON p.msg_id = m.id
`

var ErrMsgNotFound = errors.New("message not found")

func scanRows(rows *sql.Rows) ([]*domain.Message, error) {
	defer rows.Close()

	var result []*domain.Message

	for rows.Next() {
		var message domain.MessageDTO

		if err := rows.Scan(
			&message.ID,
			&message.Queue,
			&message.CreatedAt,
			&message.FinalizedAt,
			&message.Status,
			&message.StatusChangedAt,
			&message.DelayedUntil,
			&message.TimeoutAt,
			&message.Priority,
			&message.Retries,
			&message.Version,
			&message.Payload,
		); err != nil {
			return nil, err
		}

		result = append(result, domain.FromDTO(&message))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

type MessageRepository struct {
	clock  timeutils.Clock
	logger *slog.Logger
}

func NewMessageRepository(clock timeutils.Clock, logger *slog.Logger) *MessageRepository {
	return &MessageRepository{
		clock:  clock,
		logger: logger,
	}
}

func (r *MessageRepository) SaveInNewTransaction(
	ctx context.Context,
	db *sql.DB,
	msg *domain.Message,
) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("conn.BeginTx: %w", err)
	}
	defer dbutils.RollbackWithLog(tx, r.logger)

	if err := r.Save(ctx, tx, msg); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func (r *MessageRepository) Save(
	ctx context.Context,
	tx *sql.Tx,
	msg *domain.Message,
) error {
	msgDTO := msg.ToDTO()

	if msgDTO.IsNew {
		if err := r.create(ctx, tx, msgDTO); err != nil {
			return err
		}
	} else {
		if err := r.update(ctx, tx, msgDTO); err != nil {
			return err
		}
	}

	return nil
}

func (r *MessageRepository) create(
	ctx context.Context,
	tx *sql.Tx,
	msgDTO *domain.MessageDTO,
) error {
	query := `
		INSERT INTO messages (
			id, queue, created_at, finalized_at, status, status_changed_at, 
		    delayed_until, timeout_at, priority, retries, version
   		) VALUES (
			$1, $2, $3, $4, $5, $6, 
			$7, $8, $9, $10, $11
		)
    `
	if _, err := tx.ExecContext(
		ctx,
		query,
		msgDTO.ID,
		msgDTO.Queue,
		msgDTO.CreatedAt,
		msgDTO.FinalizedAt,
		msgDTO.Status,
		msgDTO.StatusChangedAt,
		msgDTO.DelayedUntil,
		msgDTO.TimeoutAt,
		msgDTO.Priority,
		msgDTO.Retries,
		msgDTO.Version,
	); err != nil {
		return err
	}

	query = `INSERT INTO message_payloads (msg_id, payload) VALUES ($1, $2)`
	if _, err := tx.ExecContext(
		ctx,
		query,
		msgDTO.ID,
		msgDTO.Payload,
	); err != nil {
		return err
	}

	return nil
}

func (r *MessageRepository) update(
	ctx context.Context,
	conn dbutils.Querier,
	msgDTO *domain.MessageDTO,
) error {
	// We know that `queue` and `created_at` never change, so we can safely skip updating them.
	query := `
		UPDATE messages
		SET finalized_at = $2,
		    status = $3, 
			status_changed_at = $4,
			delayed_until = $5,
			timeout_at = $6,
			priority = $7,
			retries = $8,
			version = version + 1
		WHERE id = $1 AND version = $9
	`
	result, err := conn.ExecContext(
		ctx,
		query,
		msgDTO.ID,
		msgDTO.FinalizedAt,
		msgDTO.Status,
		msgDTO.StatusChangedAt,
		msgDTO.DelayedUntil,
		msgDTO.TimeoutAt,
		msgDTO.Priority,
		msgDTO.Retries,
		msgDTO.Version,
	)
	if err != nil {
		return err
	}

	if count, err := result.RowsAffected(); err != nil || count != 1 {
		return errors.New("update had no effect")
	}

	return nil
}

func (r *MessageRepository) GetByID(
	ctx context.Context,
	conn dbutils.Querier,
	id string,
) (*domain.Message, error) {
	query := selectAll + `WHERE id = $1`
	rows, err := conn.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}

	messages, err := scanRows(rows)
	if err != nil {
		return nil, err
	}

	if len(messages) == 0 {
		return nil, ErrMsgNotFound
	}

	return messages[0], nil
}

func (r *MessageRepository) GetNextAvailableWithLock(
	ctx context.Context,
	conn dbutils.Querier,
	queue domain.QueueName,
	limit int,
) ([]*domain.Message, error) {
	query := selectAll + `
		WHERE queue = $1 AND status = $2
		ORDER BY priority DESC, status_changed_at ASC
		LIMIT $3
		FOR UPDATE OF m SKIP LOCKED
	`
	rows, err := conn.QueryContext(ctx, query, queue, domain.MsgStatusAvailable, limit)
	if err != nil {
		return nil, err
	}

	return scanRows(rows)
}

func (r *MessageRepository) GetProcessingToExpire(
	ctx context.Context,
	conn dbutils.Querier,
	limit int,
) ([]*domain.Message, error) {
	query := selectAll + `
		WHERE status = $1 AND timeout_at < $2
		ORDER BY timeout_at ASC
		LIMIT $3
	`
	rows, err := conn.QueryContext(ctx, query, domain.MsgStatusProcessing, r.clock.Now(), limit)
	if err != nil {
		return nil, err
	}

	return scanRows(rows)
}

func (r *MessageRepository) GetDelayedReadyToResume(
	ctx context.Context,
	conn dbutils.Querier,
	limit int,
) ([]*domain.Message, error) {
	query := selectAll + `
		WHERE status = $1 AND delayed_until < $2
		ORDER BY delayed_until ASC
		LIMIT $3
	`
	rows, err := conn.QueryContext(ctx, query, domain.MsgStatusDelayed, r.clock.Now(), limit)
	if err != nil {
		return nil, err
	}

	return scanRows(rows)
}

func (r *MessageRepository) GetFinalizedToArchive(
	ctx context.Context,
	conn dbutils.Querier,
	limit int,
) ([]*domain.Message, error) {
	query := selectAll + `
		WHERE status IN ($1, $2)
		ORDER BY finalized_at ASC
		LIMIT $3
	`
	rows, err := conn.QueryContext(
		ctx,
		query,
		domain.MsgStatusDelivered,
		domain.MsgStatusUndeliverable,
		limit,
	)
	if err != nil {
		return nil, err
	}

	return scanRows(rows)
}

func (r *MessageRepository) DeleteInNewTransaction(
	ctx context.Context,
	db *sql.DB,
	msg *domain.Message,
) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("conn.BeginTx: %w", err)
	}
	defer dbutils.RollbackWithLog(tx, r.logger)

	if err := r.Delete(ctx, tx, msg); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func (r *MessageRepository) Delete(
	ctx context.Context,
	tx *sql.Tx,
	msg *domain.Message,
) error {
	query := `DELETE FROM messages WHERE id = $1`
	if _, err := tx.ExecContext(ctx, query, msg.ID()); err != nil {
		return err
	}

	query = `DELETE FROM message_payloads WHERE msg_id = $1`
	if _, err := tx.ExecContext(ctx, query, msg.ID()); err != nil {
		return err
	}

	return nil
}
