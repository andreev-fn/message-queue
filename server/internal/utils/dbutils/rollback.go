package dbutils

import (
	"database/sql"
	"errors"
	"log/slog"
)

func RollbackWithLog(tx *sql.Tx, logger *slog.Logger) {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		logger.Error("rollback failed", "error", err)
	}
}
