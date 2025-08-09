package test

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"server/internal/appbuilder"
	"server/internal/domain"
	"server/internal/utils/testutils"
	"server/internal/utils/timeutils"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type ResponseWrapper struct {
	Success bool             `json:"success"`
	Result  *json.RawMessage `json:"result"`
	Error   *string          `json:"error"`
}

func buildTestApp(t *testing.T) (*appbuilder.App, *timeutils.StubClock) {
	t.Helper()

	if !testutils.ShouldRunIntegrationTests() {
		t.SkipNow()
	}

	clock := timeutils.NewStubClock(time.Date(2025, 6, 12, 12, 0, 0, 0, time.Local))

	app, err := appbuilder.BuildApp(&appbuilder.Config{
		DatabaseHost:     "127.0.0.1:5432",
		DatabaseUser:     "user",
		DatabasePassword: "pass",
		DatabaseName:     "queue",
	}, &appbuilder.Overrides{
		Clock: clock,
	})
	require.NoError(t, err)

	return app, clock
}

func cleanupDatabase(t *testing.T, db *sql.DB) {
	t.Helper()

	if _, err := db.Exec("DELETE FROM tasks"); err != nil {
		require.NoError(t, err)
	}
	if _, err := db.Exec("DELETE FROM task_payloads"); err != nil {
		require.NoError(t, err)
	}
	if _, err := db.Exec("DELETE FROM task_results"); err != nil {
		require.NoError(t, err)
	}
	if _, err := db.Exec("DELETE FROM archived_tasks"); err != nil {
		require.NoError(t, err)
	}
}

func createTask(t *testing.T, app *appbuilder.App, id string, kind string, priority int) *domain.Task {
	t.Helper()

	parsedID := uuid.MustParse(id)
	md5bytes := md5.Sum(parsedID[:])
	arg := hex.EncodeToString(md5bytes[:])
	task, err := domain.NewTask(
		app.Clock,
		parsedID,
		kind,
		json.RawMessage(`{"arg":"`+arg+`"}`),
		priority,
		nil,
	)
	require.NoError(t, err)

	return task
}

type NoopEventDispatcher struct{}

func (NoopEventDispatcher) Dispatch(event domain.Event) {}
