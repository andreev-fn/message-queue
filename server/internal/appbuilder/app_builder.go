package appbuilder

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"server/internal/appbuilder/requestscope"
	"server/internal/config"
	"server/internal/domain"
	"server/internal/eventbus"
	"server/internal/eventbus/postgres"
	"server/internal/routes"
	"server/internal/storage"
	"server/internal/usecases"
	"server/internal/utils/timeutils"
)

type Overrides struct {
	Clock timeutils.Clock
}

type App struct {
	Clock  timeutils.Clock
	Logger *slog.Logger
	DB     *sql.DB

	MsgRepo         *storage.MessageRepository
	ArchivedMsgRepo *storage.ArchivedMsgRepository

	EventBus *eventbus.EventBus

	RequestScopeFactory requestscope.Factory

	PublishMessages  *usecases.PublishMessages
	ReleaseMessages  *usecases.ReleaseMessages
	ConsumeMessages  *usecases.ConsumeMessages
	AckMessages      *usecases.AckMessages
	NackMessages     *usecases.NackMessages
	CheckMessages    *usecases.CheckMessages
	ArchiveMessages  *usecases.ArchiveMessages
	ExpireProcessing *usecases.ExpireProcessing
	ResumeDelayed    *usecases.ResumeDelayed

	Router *http.ServeMux
}

func BuildApp(conf *config.Config, overrides *Overrides) (*App, error) {
	if overrides == nil {
		overrides = &Overrides{}
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	var clock timeutils.Clock = timeutils.NewRealClock()
	if overrides.Clock != nil {
		clock = overrides.Clock
	}

	if conf.DatabaseType() != config.DBTypePostgres {
		return nil, errors.New("database type not supported")
	}

	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s/%s",
		conf.PostgresConfig().MustValue().Username(),
		conf.PostgresConfig().MustValue().Password(),
		conf.PostgresConfig().MustValue().Host(),
		conf.PostgresConfig().MustValue().DBName(),
	)
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	db.SetMaxIdleConns(64)
	db.SetMaxOpenConns(64)

	msgRepo := storage.NewMessageRepository(clock, logger)
	archivedMsgRepo := storage.NewArchivedMsgRepository()

	eventBus := eventbus.NewEventBus(logger, clock, postgres.NewPubSubDriver(db))

	redeliveryService := domain.NewRedeliveryService(clock, config.NewBackoffConfigProvider(conf))

	requestScopeFactory := NewRequestScopeFactory(eventBus)

	publishMessages := usecases.NewPublishMessages(logger, clock, db, msgRepo, requestScopeFactory, conf)
	releaseMessages := usecases.NewReleaseMessages(logger, clock, db, msgRepo, requestScopeFactory, conf)
	consumeMessages := usecases.NewConsumeMessages(logger, clock, db, msgRepo, eventBus, conf)
	ackMessages := usecases.NewAckMessages(clock, logger, db, msgRepo, requestScopeFactory, conf)
	nackMessages := usecases.NewNackMessages(clock, logger, db, msgRepo, redeliveryService, conf)
	checkMessages := usecases.NewCheckMessages(db, msgRepo, archivedMsgRepo, conf)
	archiveMessages := usecases.NewArchiveMessages(clock, db, msgRepo, archivedMsgRepo)
	expireProcessing := usecases.NewExpireProcessing(clock, logger, db, msgRepo, redeliveryService)
	resumeDelayed := usecases.NewResumeDelayed(clock, logger, db, msgRepo, requestScopeFactory)

	mux := http.NewServeMux()
	routes.NewPublishMessages(db, logger, publishMessages).Mount(mux)
	routes.NewReleaseMessages(db, logger, releaseMessages).Mount(mux)
	routes.NewConsumeMessages(db, logger, consumeMessages).Mount(mux)
	routes.NewAckMessages(db, logger, ackMessages).Mount(mux)
	routes.NewNackMessages(db, logger, nackMessages).Mount(mux)
	routes.NewCheckMessages(db, logger, checkMessages).Mount(mux)

	return &App{
		Clock:  clock,
		Logger: logger,
		DB:     db,

		MsgRepo:         msgRepo,
		ArchivedMsgRepo: archivedMsgRepo,

		EventBus: eventBus,

		RequestScopeFactory: requestScopeFactory,

		PublishMessages:  publishMessages,
		ReleaseMessages:  releaseMessages,
		ConsumeMessages:  consumeMessages,
		AckMessages:      ackMessages,
		NackMessages:     nackMessages,
		CheckMessages:    checkMessages,
		ArchiveMessages:  archiveMessages,
		ExpireProcessing: expireProcessing,
		ResumeDelayed:    resumeDelayed,

		Router: mux,
	}, nil
}
