package e2eutils

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"server/internal/appbuilder"
	"server/internal/utils/testutils"
	"server/internal/utils/timeutils"
)

type ResponseWrapper struct {
	Success bool             `json:"success"`
	Result  *json.RawMessage `json:"result"`
	Error   *string          `json:"error"`
}

func Prepare(t *testing.T) (*appbuilder.App, *timeutils.StubClock) {
	t.Helper()

	if !testutils.ShouldRunIntegrationTests() {
		t.SkipNow()
	}

	app, clock := BuildTestApp(t)

	CleanupDatabase(t, app.DB)

	return app, clock
}

func BuildTestApp(t *testing.T) (*appbuilder.App, *timeutils.StubClock) {
	t.Helper()

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

func CleanupDatabase(t *testing.T, db *sql.DB) {
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
