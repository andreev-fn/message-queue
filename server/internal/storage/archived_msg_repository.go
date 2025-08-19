package storage

import (
	"context"
	"errors"

	"server/internal/domain"
	"server/internal/utils/dbutils"
)

var ErrArchivedMsgNotFound = errors.New("archived message not found")

type ArchivedMsgRepository struct{}

func NewArchivedMsgRepository() *ArchivedMsgRepository {
	return &ArchivedMsgRepository{}
}

func (r *ArchivedMsgRepository) Upsert(
	ctx context.Context,
	conn dbutils.Querier,
	msg *domain.ArchivedMsg,
) error {
	msgDTO := msg.ToDTO()

	query := `
		INSERT INTO archived_messages (
			id, queue, created_at, finalized_at, status, priority, retries, payload
   		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET 
		    queue = $2, 
		    created_at = $3, 
		    finalized_at = $4, 
		    status = $5, 
		    priority = $6, 
		    retries = $7, 
		    payload = $8
    `
	if _, err := conn.ExecContext(
		ctx,
		query,
		msgDTO.ID,
		msgDTO.Queue,
		msgDTO.CreatedAt,
		msgDTO.FinalizedAt,
		msgDTO.Status,
		msgDTO.Priority,
		msgDTO.Retries,
		msgDTO.Payload,
	); err != nil {
		return err
	}

	return nil
}

func (r *ArchivedMsgRepository) GetByID(
	ctx context.Context,
	conn dbutils.Querier,
	id string,
) (*domain.ArchivedMsg, error) {
	query := `
		SELECT id, queue, created_at, finalized_at, status, priority, retries, payload
		FROM archived_messages
		WHERE id = $1
	`
	rows, err := conn.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []*domain.ArchivedMsg

	for rows.Next() {
		var msg domain.ArchivedMsgDTO

		if err := rows.Scan(
			&msg.ID,
			&msg.Queue,
			&msg.CreatedAt,
			&msg.FinalizedAt,
			&msg.Status,
			&msg.Priority,
			&msg.Retries,
			&msg.Payload,
		); err != nil {
			return nil, err
		}

		result = append(result, domain.ArchivedMsgFromDTO(&msg))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, ErrArchivedMsgNotFound
	}

	return result[0], nil
}

func (r *ArchivedMsgRepository) Delete(
	ctx context.Context,
	conn dbutils.Querier,
	msg *domain.ArchivedMsg,
) error {
	query := `DELETE FROM archived_messages WHERE id = $1`
	if _, err := conn.ExecContext(ctx, query, msg.ID()); err != nil {
		return err
	}

	return nil
}
