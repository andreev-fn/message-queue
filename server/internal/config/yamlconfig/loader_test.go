package yamlconfig

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"server/internal/config"
	"server/internal/domain"
)

func TestLoadFromFile_full(t *testing.T) {
	t.Setenv("DB_PASS", "secret")
	t.Setenv("BATCH_SIZE_MAX", "122")

	cfg, err := LoadFromFile("testdata/config.full.yaml")
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
	for _, qName := range []string{"queue1", "queue2"} {
		q, err := cfg.GetQueueConfig(domain.UnsafeQueueName(qName))
		require.NoError(t, err)

		require.Equal(t, 5*time.Minute, q.ProcessingTimeout())
		require.True(t, q.IsDeadLetteringOn())

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
	}
}

func TestLoadFromFile_minimal(t *testing.T) {
	cfg, err := LoadFromFile("testdata/config.minimal.yaml")
	require.NoError(t, err, "expected config to load without error")
	require.NotNil(t, cfg)

	// DB
	require.Equal(t, config.DBTypePostgres, cfg.DatabaseType())
	require.True(t, cfg.PostgresConfig().IsSet())
	require.Equal(t, "127.0.0.1:5432", cfg.PostgresConfig().MustValue().Host())
	require.Equal(t, "queue", cfg.PostgresConfig().MustValue().DBName())
	require.Equal(t, "user", cfg.PostgresConfig().MustValue().Username())
	require.Equal(t, "", cfg.PostgresConfig().MustValue().Password())

	// App
	require.Equal(t, config.DefaultBatchSizeLimit, cfg.BatchSizeLimit())

	// Queue
	q, err := cfg.GetQueueConfig(domain.UnsafeQueueName("queue1"))
	require.NoError(t, err)

	require.Equal(t, 5*time.Minute, q.ProcessingTimeout())
	require.True(t, q.IsDeadLetteringOn())

	// Backoff
	require.Equal(t, config.DefaultBackoffEnabled, q.Backoff().IsSet())
	if config.DefaultBackoffEnabled {
		require.Equal(t, config.DefaultBackoffShape(), q.Backoff().MustValue().Shape())
		require.True(t, q.Backoff().MustValue().MaxAttempts().IsSet())
		require.Equal(t, config.DefaultBackoffMaxAttempts, q.Backoff().MustValue().MaxAttempts().MustValue())
	}
}

func TestLoadFromFile_disabled(t *testing.T) {
	cfg, err := LoadFromFile("testdata/config.disabled.yaml")
	require.NoError(t, err, "expected config to load without error")
	require.NotNil(t, cfg)

	// DB
	require.Equal(t, config.DBTypePostgres, cfg.DatabaseType())
	require.True(t, cfg.PostgresConfig().IsSet())
	require.Equal(t, "127.0.0.1:5432", cfg.PostgresConfig().MustValue().Host())
	require.Equal(t, "queue", cfg.PostgresConfig().MustValue().DBName())
	require.Equal(t, "user", cfg.PostgresConfig().MustValue().Username())
	require.Equal(t, "", cfg.PostgresConfig().MustValue().Password())

	// App
	require.Equal(t, config.DefaultBatchSizeLimit, cfg.BatchSizeLimit())

	// Queue
	q, err := cfg.GetQueueConfig(domain.UnsafeQueueName("queue1"))
	require.NoError(t, err)

	require.Equal(t, 5*time.Minute, q.ProcessingTimeout())
	require.False(t, q.IsDeadLetteringOn())

	// Backoff
	require.False(t, q.Backoff().IsSet())
}

func TestLoadFromFile_custom(t *testing.T) {
	cfg, err := LoadFromFile("testdata/config.custom.yaml")
	require.NoError(t, err, "expected config to load without error")
	require.NotNil(t, cfg)

	// DB
	require.Equal(t, config.DBTypePostgres, cfg.DatabaseType())
	require.True(t, cfg.PostgresConfig().IsSet())
	require.Equal(t, "127.0.0.1:5432", cfg.PostgresConfig().MustValue().Host())
	require.Equal(t, "queue", cfg.PostgresConfig().MustValue().DBName())
	require.Equal(t, "user", cfg.PostgresConfig().MustValue().Username())
	require.Equal(t, "", cfg.PostgresConfig().MustValue().Password())

	// App
	require.Equal(t, config.DefaultBatchSizeLimit, cfg.BatchSizeLimit())

	// Queue
	q, err := cfg.GetQueueConfig(domain.UnsafeQueueName("queue1"))
	require.NoError(t, err)

	require.Equal(t, 5*time.Minute, q.ProcessingTimeout())
	require.True(t, q.IsDeadLetteringOn())

	// Backoff
	require.True(t, q.Backoff().IsSet())
	require.Equal(t, config.DefaultBackoffShape(), q.Backoff().MustValue().Shape())
	require.False(t, q.Backoff().MustValue().MaxAttempts().IsSet())
}

func TestLoadFromFile_DirectConfigOfDLQNotAllowed(t *testing.T) {
	_, err := LoadFromFile("testdata/config.err.dlq.yaml")
	require.ErrorContains(t, err, "manual configuration of DL queues is not allowed")
}
