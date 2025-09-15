package yamlconfig

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"server/internal/config"
)

func TestLoad(t *testing.T) {
	t.Setenv("DB_PASS", "secret")
	t.Setenv("BATCH_SIZE_MAX", "122")

	cfg, err := Load("testdata/config.yaml")
	require.NoError(t, err, "expected config to load without error")
	require.NotNil(t, cfg)

	// DB
	require.Equal(t, config.DBTypePostgres, cfg.DatabaseType())
	require.True(t, cfg.PostgresConfig().IsSet())
	require.Equal(t, "127.0.0.1:5432", cfg.PostgresConfig().MustValue().Host())
	require.Equal(t, "queue", cfg.PostgresConfig().MustValue().DBName())
	require.Equal(t, "user", cfg.PostgresConfig().MustValue().Username())
	require.Equal(t, "secret", cfg.PostgresConfig().MustValue().Password())

	// App
	require.Equal(t, 122, cfg.BatchSizeLimit())

	// Queues
	require.True(t, cfg.IsQueueDefined("queue1"))
	require.True(t, cfg.IsQueueDefined("queue2"))

	for _, qName := range []string{"queue1", "queue2"} {
		q := cfg.QueueConfig(qName)
		// Backoff
		require.True(t, q.Backoff().IsSet())
		require.Equal(t, []time.Duration{
			30 * time.Second,
			1 * time.Minute,
			2 * time.Minute,
			5 * time.Minute,
		}, q.Backoff().MustValue().Shape())
		require.True(t, q.Backoff().MustValue().MaxAttempts().IsSet())
		require.Equal(t, 10, q.Backoff().MustValue().MaxAttempts().MustValue())

		// Timeouts and retention
		require.Equal(t, 5*time.Minute, q.ProcessingTimeout())
		require.True(t, q.RetentionPeriod().IsSet())
		require.Equal(t, 720*time.Hour, q.RetentionPeriod().MustValue())
	}
}
