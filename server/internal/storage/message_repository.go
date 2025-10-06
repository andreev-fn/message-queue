package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"server/internal/domain"
	"server/internal/utils/dbutils"
	"server/internal/utils/timeutils"
)

const selectAll = `
	SELECT 
		m.id, m.queue, m.created_at, m.finalized_at, m.status, m.status_changed_at,
		m.delayed_until, m.timeout_at, m.priority, m.retries, m.generation ,m.version,
		p.payload
	FROM messages m
	LEFT JOIN message_payloads p ON p.msg_id = m.id
`

var ErrMsgNotFound = errors.New("message not found")

func scanRows(rows *sql.Rows) ([]*domain.MessageDTO, error) {
	defer rows.Close()

	var result []*domain.MessageDTO

	for rows.Next() {
		var dto domain.MessageDTO

		if err := rows.Scan(
			&dto.ID,
			&dto.Queue,
			&dto.CreatedAt,
			&dto.FinalizedAt,
			&dto.Status,
			&dto.StatusChangedAt,
			&dto.DelayedUntil,
			&dto.TimeoutAt,
			&dto.Priority,
			&dto.Retries,
			&dto.Generation,
			&dto.Version,
			&dto.Payload,
		); err != nil {
			return nil, err
		}

		result = append(result, &dto)
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

	for _, chapter := range msgDTO.History {
		if !chapter.IsNew {
			continue
		}
		if err := r.createHistoryChapter(ctx, tx, chapter); err != nil {
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
		    delayed_until, timeout_at, priority, retries, generation, version
   		) VALUES (
			$1, $2, $3, $4, $5, $6, 
			$7, $8, $9, $10, $11, $12
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
		msgDTO.Generation,
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
	// Field `created_at` never change, there is no need to update it.
	query := `
		UPDATE messages
		SET queue = $2,
		    finalized_at = $3,
		    status = $4, 
			status_changed_at = $5,
			delayed_until = $6,
			timeout_at = $7,
			priority = $8,
			retries = $9,
			generation = $10,
			version = version + 1
		WHERE id = $1 AND version = $11
	`
	result, err := conn.ExecContext(
		ctx,
		query,
		msgDTO.ID,
		msgDTO.Queue,
		msgDTO.FinalizedAt,
		msgDTO.Status,
		msgDTO.StatusChangedAt,
		msgDTO.DelayedUntil,
		msgDTO.TimeoutAt,
		msgDTO.Priority,
		msgDTO.Retries,
		msgDTO.Generation,
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

func (r *MessageRepository) createHistoryChapter(
	ctx context.Context,
	tx *sql.Tx,
	chapterDTO *domain.MessageChapterDTO,
) error {
	query := `
		INSERT INTO message_history (
			msg_id, generation, queue, redirected_at, priority, retries
   		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
    `
	if _, err := tx.ExecContext(
		ctx,
		query,
		chapterDTO.MsgID,
		chapterDTO.Generation,
		chapterDTO.Queue,
		chapterDTO.RedirectedAt,
		chapterDTO.Priority,
		chapterDTO.Retries,
	); err != nil {
		return err
	}

	return nil
}

func (r *MessageRepository) GetByID(
	ctx context.Context,
	conn dbutils.Querier,
	id string,
) (*domain.Message, error) {
	return r.getByID(ctx, conn, id, false)
}

func (r *MessageRepository) GetByIDWithHistory(
	ctx context.Context,
	db *sql.DB,
	id string,
) (*domain.Message, error) {
	// Must run in REPEATABLE READ, otherwise history could change between message and history queries.

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, fmt.Errorf("conn.BeginTx: %w", err)
	}
	defer dbutils.RollbackWithLog(tx, r.logger)

	msg, err := r.getByID(ctx, tx, id, true)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (r *MessageRepository) getByID(
	ctx context.Context,
	conn dbutils.Querier,
	id string,
	withHistory bool,
) (*domain.Message, error) {
	query := selectAll + `WHERE id = $1`
	rows, err := conn.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}

	dtos, err := scanRows(rows)
	if err != nil {
		return nil, err
	}

	if len(dtos) == 0 {
		return nil, ErrMsgNotFound
	}

	if withHistory {
		history, err := r.getHistory(ctx, conn, []string{id})
		if err != nil {
			return nil, err
		}

		dtos[0].History = history[id]
	}

	return domain.FromDTO(dtos[0]), nil
}

func (r *MessageRepository) getHistory(
	ctx context.Context,
	conn dbutils.Querier,
	msgIDs []string,
) (map[string][]*domain.MessageChapterDTO, error) {
	if len(msgIDs) == 0 {
		return map[string][]*domain.MessageChapterDTO{}, nil
	}

	placeholders := make([]string, len(msgIDs))
	args := make([]interface{}, len(msgIDs))
	for i, msgID := range msgIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = msgID
	}

	query := fmt.Sprintf(`
		SELECT msg_id, generation, queue, redirected_at, priority, retries 
		FROM message_history WHERE msg_id IN (%s) ORDER BY msg_id, generation
	`, strings.Join(placeholders, ", "))

	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make(map[string][]*domain.MessageChapterDTO, len(msgIDs))
	for _, msgID := range msgIDs {
		result[msgID] = make([]*domain.MessageChapterDTO, 0)
	}

	for rows.Next() {
		var dto domain.MessageChapterDTO

		if err := rows.Scan(
			&dto.MsgID,
			&dto.Generation,
			&dto.Queue,
			&dto.RedirectedAt,
			&dto.Priority,
			&dto.Retries,
		); err != nil {
			return nil, err
		}

		msgIDStr := dto.MsgID.String()
		result[msgIDStr] = append(result[msgIDStr], &dto)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
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

	return mapToMessages(scanRows(rows))
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

	return mapToMessages(scanRows(rows))
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

	return mapToMessages(scanRows(rows))
}

func (r *MessageRepository) GetFinalizedToArchive(
	ctx context.Context,
	conn dbutils.Querier,
	limit int,
) ([]*domain.Message, error) {
	// It's safe here to get messages and history in separate transactions,
	// because messages are in final statuses and history won't change.

	query := selectAll + `
		WHERE status IN ($1, $2)
		ORDER BY finalized_at ASC
		LIMIT $3
	`
	rows, err := conn.QueryContext(
		ctx,
		query,
		domain.MsgStatusDelivered,
		domain.MsgStatusDropped,
		limit,
	)
	if err != nil {
		return nil, err
	}

	dtos, err := scanRows(rows)
	if err != nil {
		return nil, err
	}

	msgIDs := make([]string, 0, len(dtos))
	for _, dto := range dtos {
		msgIDs = append(msgIDs, dto.ID.String())
	}

	history, err := r.getHistory(ctx, conn, msgIDs)
	if err != nil {
		return nil, fmt.Errorf("getHistory: %w", err)
	}

	for i, dto := range dtos {
		dtos[i].History = history[dto.ID.String()]
	}

	return mapToMessages(dtos, nil)
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

	query = `DELETE FROM message_history WHERE msg_id = $1`
	if _, err := tx.ExecContext(ctx, query, msg.ID()); err != nil {
		return err
	}

	return nil
}

func mapToMessages(dtos []*domain.MessageDTO, err error) ([]*domain.Message, error) {
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Message, 0, len(dtos))
	for _, dto := range dtos {
		result = append(result, domain.FromDTO(dto))
	}
	return result, nil
}
