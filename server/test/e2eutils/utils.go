package e2eutils

import (
	"database/sql"
	"testing"
	"time"

	"server/internal/appbuilder"
	"server/internal/config"
	"server/internal/domain"
	"server/internal/utils/opt"
	"server/internal/utils/timeutils"
	"server/pkg/httpclient"
)

func PrepareHTTPClient(t *testing.T, app *appbuilder.App) *httpclient.Client {
	return httpclient.NewClient("/", NewHTTPTestDoer(t, app.Router))
}

func Prepare() *appbuilder.App {
	app := BuildTestApp(CreateTestConfig())
	CleanupDatabase(app.DB)
	return app
}

func CreateTestConfig() *config.Config {
	pgConf, err := config.NewPostgresConfig(
		"127.0.0.1:5432",
		"queue",
		"user",
		"pass",
	)
	if err != nil {
		panic(err)
	}

	backoffConfig, err := domain.NewBackoffConfig(
		[]time.Duration{time.Second * 30},
		opt.Some(config.DefaultBackoffMaxAttempts),
	)
	if err != nil {
		panic(err)
	}

	queueConfig, err := domain.NewQueueConfig(
		opt.Some(backoffConfig),
		time.Minute*5,
	)
	if err != nil {
		panic(err)
	}

	conf, err := config.NewConfig(
		opt.Some(pgConf),
		config.DefaultBatchSizeLimit,
		map[domain.QueueName]*domain.QueueConfig{
			domain.UnsafeQueueName("test"):        queueConfig,
			domain.UnsafeQueueName("test.result"): queueConfig,
			domain.UnsafeQueueName("all_results"): queueConfig,
		},
	)
	if err != nil {
		panic(err)
	}

	return conf
}

func BuildTestApp(conf *config.Config) *appbuilder.App {
	clock := timeutils.NewStubClock(time.Date(2025, 6, 12, 12, 0, 0, 0, time.Local))

	app, err := appbuilder.BuildApp(conf, &appbuilder.Overrides{
		Clock: clock,
	})
	if err != nil {
		panic(err)
	}

	return app
}

func AdvanceClock(app *appbuilder.App, dur time.Duration) {
	app.Clock.(*timeutils.StubClock).Set(app.Clock.Now().Add(dur))
}

func CleanupDatabase(db *sql.DB) {
	if _, err := db.Exec("DELETE FROM messages"); err != nil {
		panic(err)
	}
	if _, err := db.Exec("DELETE FROM message_payloads"); err != nil {
		panic(err)
	}
	if _, err := db.Exec("DELETE FROM message_history"); err != nil {
		panic(err)
	}
	if _, err := db.Exec("DELETE FROM archived_messages"); err != nil {
		panic(err)
	}
}
