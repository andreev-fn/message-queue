package storage

import (
	"context"
	"errors"
	"server/internal/domain"
	"server/internal/utils/dbutils"
)

var ErrArchivedTaskNotFound = errors.New("archived task not found")

type ArchivedTaskRepository struct{}

func NewArchivedTaskRepository() *ArchivedTaskRepository {
	return &ArchivedTaskRepository{}
}

func (r *ArchivedTaskRepository) Upsert(
	ctx context.Context,
	conn dbutils.Querier,
	task *domain.ArchivedTask,
) error {
	taskDTO := task.ToDTO()

	query := `
		INSERT INTO archived_tasks (
			id, kind, created_at, finalized_at, status, priority, retries, payload, result
   		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET 
		    kind = $2, 
		    created_at = $3, 
		    finalized_at = $4, 
		    status = $5, 
		    priority = $6, 
		    retries = $7, 
		    payload = $8, 
		    result = $9
    `
	if _, err := conn.ExecContext(
		ctx,
		query,
		taskDTO.ID,
		taskDTO.Kind,
		taskDTO.CreatedAt,
		taskDTO.FinalizedAt,
		taskDTO.Status,
		taskDTO.Priority,
		taskDTO.Retries,
		taskDTO.Payload,
		taskDTO.Result,
	); err != nil {
		return err
	}

	return nil
}

func (r *ArchivedTaskRepository) GetTaskByID(
	ctx context.Context,
	conn dbutils.Querier,
	id string,
) (*domain.ArchivedTask, error) {
	query := `
		SELECT t.id, t.kind, t.created_at, t.finalized_at, t.status, t.priority, t.retries, t.payload, t.result
		FROM archived_tasks t
		WHERE id = $1
	`
	rows, err := conn.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []*domain.ArchivedTask

	for rows.Next() {
		var task domain.ArchivedTaskDTO

		if err := rows.Scan(
			&task.ID,
			&task.Kind,
			&task.CreatedAt,
			&task.FinalizedAt,
			&task.Status,
			&task.Priority,
			&task.Retries,
			&task.Payload,
			&task.Result,
		); err != nil {
			return nil, err
		}

		result = append(result, domain.ArchivedTaskFromDTO(&task))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, ErrArchivedTaskNotFound
	}

	return result[0], nil
}

func (r *ArchivedTaskRepository) Delete(
	ctx context.Context,
	conn dbutils.Querier,
	task *domain.ArchivedTask,
) error {
	query := `DELETE FROM archived_tasks WHERE id = $1`
	if _, err := conn.ExecContext(ctx, query, task.ID()); err != nil {
		return err
	}

	return nil
}
