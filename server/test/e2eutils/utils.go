package e2eutils

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"server/internal/appbuilder"
	"server/internal/config"
	"server/internal/domain"
	"server/internal/utils/opt"
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

	app, clock := BuildTestApp(t, CreateTestConfig(t))

	CleanupDatabase(t, app.DB)

	return app, clock
}

func CreateTestConfig(t *testing.T) *config.Config {
	t.Helper()

	pgConf, err := config.NewPostgresConfig(
		"127.0.0.1:5432",
		"queue",
		"user",
		"pass",
	)
	require.NoError(t, err)

	backoffConfig, err := domain.NewBackoffConfig(
		[]time.Duration{time.Second * 30},
		opt.Some(config.DefaultBackoffMaxAttempts),
	)
	require.NoError(t, err)

	queueConfig, err := domain.NewQueueConfig(
		opt.Some(backoffConfig),
		time.Minute*5,
	)
	require.NoError(t, err)

	conf, err := config.NewConfig(
		opt.Some(pgConf),
		config.DefaultBatchSizeLimit,
		map[string]*domain.QueueConfig{
			"test":        queueConfig,
			"test.result": queueConfig,
		},
	)
	require.NoError(t, err)

	return conf
}

func BuildTestApp(t *testing.T, conf *config.Config) (*appbuilder.App, *timeutils.StubClock) {
	t.Helper()

	clock := timeutils.NewStubClock(time.Date(2025, 6, 12, 12, 0, 0, 0, time.Local))

	app, err := appbuilder.BuildApp(conf, &appbuilder.Overrides{
		Clock: clock,
	})
	require.NoError(t, err)

	return app, clock
}

func CleanupDatabase(t *testing.T, db *sql.DB) {
	t.Helper()

	if _, err := db.Exec("DELETE FROM messages"); err != nil {
		require.NoError(t, err)
	}
	if _, err := db.Exec("DELETE FROM message_payloads"); err != nil {
		require.NoError(t, err)
	}
	if _, err := db.Exec("DELETE FROM archived_messages"); err != nil {
		require.NoError(t, err)
	}
}
