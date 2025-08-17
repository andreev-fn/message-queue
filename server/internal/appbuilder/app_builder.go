package appbuilder

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"server/internal/appbuilder/requestscope"
	"server/internal/eventbus"
	"server/internal/eventbus/postgres"
	"server/internal/routes"
	"server/internal/storage"
	"server/internal/usecases"
	"server/internal/utils/timeutils"
)

type Config struct {
	DatabaseHost     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
}

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

	CreateMessage    *usecases.CreateMessage
	ConfirmMessage   *usecases.ConfirmMessage
	TakeWork         *usecases.TakeWork
	FinishWork       *usecases.FinishWork
	CheckMessage     *usecases.CheckMessage
	ArchiveMessages  *usecases.ArchiveMessages
	ExpireProcessing *usecases.ExpireProcessing
	ResumeDelayed    *usecases.ResumeDelayed

	Router *http.ServeMux
}

func BuildApp(conf *Config, overrides *Overrides) (*App, error) {
	if overrides == nil {
		overrides = &Overrides{}
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	var clock timeutils.Clock = timeutils.NewRealClock()
	if overrides.Clock != nil {
		clock = overrides.Clock
	}

	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s/%s",
		conf.DatabaseUser,
		conf.DatabasePassword,
		conf.DatabaseHost,
		conf.DatabaseName,
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

	requestScopeFactory := NewRequestScopeFactory(eventBus)

	createMessage := usecases.NewCreateMessage(logger, clock, db, msgRepo, requestScopeFactory)
	confirmMessage := usecases.NewConfirmMessage(logger, clock, db, msgRepo, requestScopeFactory)
	takeWork := usecases.NewTakeWork(logger, clock, db, msgRepo, eventBus)
	finishWork := usecases.NewFinishWork(clock, logger, db, msgRepo)
	checkMessage := usecases.NewCheckMessage(db, msgRepo, archivedMsgRepo)
	archiveMessages := usecases.NewArchiveMessages(clock, db, msgRepo, archivedMsgRepo)
	expireProcessing := usecases.NewExpireProcessing(clock, logger, db, msgRepo)
	resumeDelayed := usecases.NewResumeDelayed(clock, logger, db, msgRepo, requestScopeFactory)

	mux := http.NewServeMux()
	routes.NewCreateMessage(db, logger, createMessage).Mount(mux)
	routes.NewConfirmMessage(db, logger, confirmMessage).Mount(mux)
	routes.NewTakeWork(db, logger, takeWork).Mount(mux)
	routes.NewFinishWork(db, logger, finishWork).Mount(mux)
	routes.NewCheckMessage(db, logger, checkMessage).Mount(mux)

	return &App{
		Clock:  clock,
		Logger: logger,
		DB:     db,

		MsgRepo:         msgRepo,
		ArchivedMsgRepo: archivedMsgRepo,

		EventBus: eventBus,

		RequestScopeFactory: requestScopeFactory,

		CreateMessage:    createMessage,
		ConfirmMessage:   confirmMessage,
		TakeWork:         takeWork,
		FinishWork:       finishWork,
		CheckMessage:     checkMessage,
		ArchiveMessages:  archiveMessages,
		ExpireProcessing: expireProcessing,
		ResumeDelayed:    resumeDelayed,

		Router: mux,
	}, nil
}
