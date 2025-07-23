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
	"strings"
)

const selectAll = `
	SELECT 
		t.id, t.kind, t.created_at, t.finalized_at, t.status, t.status_changed_at,
		t.delayed_until, t.timeout_at, t.priority, t.retries, t.version,
		p.payload,
		r.result
	FROM tasks t
	LEFT JOIN task_payloads p ON p.task_id = t.id
	LEFT JOIN task_results r ON r.task_id = t.id
`

var ErrTaskNotFound = errors.New("task not found")

func scanRows(rows *sql.Rows) ([]*domain.Task, error) {
	defer rows.Close()

	var result []*domain.Task

	for rows.Next() {
		var task domain.TaskDTO

		if err := rows.Scan(
			&task.ID,
			&task.Kind,
			&task.CreatedAt,
			&task.FinalizedAt,
			&task.Status,
			&task.StatusChangedAt,
			&task.DelayedUntil,
			&task.TimeoutAt,
			&task.Priority,
			&task.Retries,
			&task.Version,
			&task.Payload,
			&task.Result,
		); err != nil {
			return nil, err
		}

		result = append(result, domain.FromDTO(&task))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

type TaskRepository struct {
	clock  timeutils.Clock
	logger *slog.Logger
}

func NewTaskRepository(clock timeutils.Clock, logger *slog.Logger) *TaskRepository {
	return &TaskRepository{
		clock:  clock,
		logger: logger,
	}
}

func (r *TaskRepository) SaveInNewTransaction(
	ctx context.Context,
	db *sql.DB,
	task *domain.Task,
) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("conn.BeginTx: %w", err)
	}
	defer dbutils.RollbackWithLog(tx, r.logger)

	if err := r.Save(ctx, tx, task); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func (r *TaskRepository) Save(
	ctx context.Context,
	tx *sql.Tx,
	task *domain.Task,
) error {
	taskDTO := task.ToDTO()

	if taskDTO.IsNew {
		if err := r.create(ctx, tx, taskDTO); err != nil {
			return err
		}
	} else {
		if err := r.update(ctx, tx, taskDTO); err != nil {
			return err
		}
	}

	if taskDTO.IsResultNew {
		query := `INSERT INTO task_results (task_id, result) VALUES ($1, $2)`
		if _, err := tx.ExecContext(
			ctx,
			query,
			taskDTO.ID,
			taskDTO.Result,
		); err != nil {
			return err
		}
	}

	return nil
}

func (r *TaskRepository) create(
	ctx context.Context,
	tx *sql.Tx,
	taskDTO *domain.TaskDTO,
) error {
	query := `
		INSERT INTO tasks (
			id, kind, created_at, finalized_at, status, status_changed_at, 
		    delayed_until, timeout_at, priority, retries, version
   		) VALUES (
			$1, $2, $3, $4, $5, $6, 
			$7, $8, $9, $10, $11
		)
    `
	if _, err := tx.ExecContext(
		ctx,
		query,
		taskDTO.ID,
		taskDTO.Kind,
		taskDTO.CreatedAt,
		taskDTO.FinalizedAt,
		taskDTO.Status,
		taskDTO.StatusChangedAt,
		taskDTO.DelayedUntil,
		taskDTO.TimeoutAt,
		taskDTO.Priority,
		taskDTO.Retries,
		taskDTO.Version,
	); err != nil {
		return err
	}

	query = `INSERT INTO task_payloads (task_id, payload) VALUES ($1, $2)`
	if _, err := tx.ExecContext(
		ctx,
		query,
		taskDTO.ID,
		taskDTO.Payload,
	); err != nil {
		return err
	}

	return nil
}

func (r *TaskRepository) update(
	ctx context.Context,
	conn dbutils.Querier,
	taskDTO *domain.TaskDTO,
) error {
	// We know that `kind` and `created_at` never change, so we can safely skip updating them.
	query := `
		UPDATE tasks
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
		taskDTO.ID,
		taskDTO.FinalizedAt,
		taskDTO.Status,
		taskDTO.StatusChangedAt,
		taskDTO.DelayedUntil,
		taskDTO.TimeoutAt,
		taskDTO.Priority,
		taskDTO.Retries,
		taskDTO.Version,
	)
	if err != nil {
		return err
	}

	if count, err := result.RowsAffected(); err != nil || count != 1 {
		return errors.New("update had no effect")
	}

	return nil
}

func (r *TaskRepository) GetTaskByID(
	ctx context.Context,
	conn dbutils.Querier,
	id string,
) (*domain.Task, error) {
	query := selectAll + `WHERE id = $1`
	rows, err := conn.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}

	tasks, err := scanRows(rows)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, ErrTaskNotFound
	}

	return tasks[0], nil
}

func (r *TaskRepository) GetReadyWithLock(
	ctx context.Context,
	conn dbutils.Querier,
	kinds []string,
	limit int,
) ([]*domain.Task, error) {
	if len(kinds) == 0 {
		return []*domain.Task{}, nil
	}

	query := selectAll + `
		WHERE kind IN (:kinds) AND status = $1
		ORDER BY priority DESC, status_changed_at ASC
		LIMIT $2
		FOR UPDATE OF t SKIP LOCKED
	`
	query = strings.ReplaceAll(query, ":kinds", "'"+strings.Join(kinds, "','")+"'")
	rows, err := conn.QueryContext(ctx, query, domain.TaskStatusReady, limit)
	if err != nil {
		return nil, err
	}

	return scanRows(rows)
}

func (r *TaskRepository) GetProcessingToExpire(
	ctx context.Context,
	conn dbutils.Querier,
	limit int,
) ([]*domain.Task, error) {
	query := selectAll + `
		WHERE status = $1 AND timeout_at < $2
		ORDER BY timeout_at ASC
		LIMIT $3
	`
	rows, err := conn.QueryContext(ctx, query, domain.TaskStatusProcessing, r.clock.Now(), limit)
	if err != nil {
		return nil, err
	}

	return scanRows(rows)
}

func (r *TaskRepository) GetDelayedReadyToResume(
	ctx context.Context,
	conn dbutils.Querier,
	limit int,
) ([]*domain.Task, error) {
	query := selectAll + `
		WHERE status = $1 AND delayed_until < $2
		ORDER BY delayed_until ASC
		LIMIT $3
	`
	rows, err := conn.QueryContext(ctx, query, domain.TaskStatusDelayed, r.clock.Now(), limit)
	if err != nil {
		return nil, err
	}

	return scanRows(rows)
}

func (r *TaskRepository) GetTasksToArchive(
	ctx context.Context,
	conn dbutils.Querier,
	limit int,
) ([]*domain.Task, error) {
	query := selectAll + `
		WHERE status IN ($1, $2)
		ORDER BY finalized_at ASC
		LIMIT $3
	`
	rows, err := conn.QueryContext(
		ctx,
		query,
		domain.TaskStatusCompleted,
		domain.TaskStatusFailed,
		limit,
	)
	if err != nil {
		return nil, err
	}

	return scanRows(rows)
}

func (r *TaskRepository) DeleteInNewTransaction(
	ctx context.Context,
	db *sql.DB,
	task *domain.Task,
) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("conn.BeginTx: %w", err)
	}
	defer dbutils.RollbackWithLog(tx, r.logger)

	if err := r.Delete(ctx, tx, task); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func (r *TaskRepository) Delete(
	ctx context.Context,
	tx *sql.Tx,
	task *domain.Task,
) error {
	query := `DELETE FROM tasks WHERE id = $1`
	if _, err := tx.ExecContext(ctx, query, task.ID()); err != nil {
		return err
	}

	query = `DELETE FROM task_payloads WHERE task_id = $1`
	if _, err := tx.ExecContext(ctx, query, task.ID()); err != nil {
		return err
	}

	query = `DELETE FROM task_results WHERE task_id = $1`
	if _, err := tx.ExecContext(ctx, query, task.ID()); err != nil {
		return err
	}

	return nil
}
