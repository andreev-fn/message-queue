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

	TaskRepo         *storage.TaskRepository
	ArchivedTaskRepo *storage.ArchivedTaskRepository

	EventBus *eventbus.EventBus

	RequestScopeFactory requestscope.Factory

	CreateTask       *usecases.CreateTask
	ConfirmTask      *usecases.ConfirmTask
	TakeWork         *usecases.TakeWork
	FinishWork       *usecases.FinishWork
	CheckTask        *usecases.CheckTask
	ArchiveTasks     *usecases.ArchiveTasks
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

	taskRepo := storage.NewTaskRepository(clock, logger)
	archivedTaskRepo := storage.NewArchivedTaskRepository()

	eventBus := eventbus.NewEventBus(logger, clock, postgres.NewPubSubDriver(db))

	requestScopeFactory := NewRequestScopeFactory(eventBus)

	createTask := usecases.NewCreateTask(logger, clock, db, taskRepo, requestScopeFactory)
	confirmTask := usecases.NewConfirmTask(logger, clock, db, taskRepo, requestScopeFactory)
	takeWork := usecases.NewTakeWork(logger, clock, db, taskRepo, eventBus)
	finishWork := usecases.NewFinishWork(clock, logger, db, taskRepo)
	checkTask := usecases.NewCheckTask(db, taskRepo, archivedTaskRepo)
	archiveTasks := usecases.NewArchiveTasks(clock, db, taskRepo, archivedTaskRepo)
	expireProcessing := usecases.NewExpireProcessing(clock, logger, db, taskRepo)
	resumeDelayed := usecases.NewResumeDelayed(clock, logger, db, taskRepo, requestScopeFactory)

	mux := http.NewServeMux()
	routes.NewCreateTask(db, logger, createTask).Mount(mux)
	routes.NewConfirmTask(db, logger, confirmTask).Mount(mux)
	routes.NewTakeWork(db, logger, takeWork).Mount(mux)
	routes.NewFinishWork(db, logger, finishWork).Mount(mux)
	routes.NewCheckTask(db, logger, checkTask).Mount(mux)

	return &App{
		Clock:  clock,
		Logger: logger,
		DB:     db,

		TaskRepo:         taskRepo,
		ArchivedTaskRepo: archivedTaskRepo,

		EventBus: eventBus,

		RequestScopeFactory: requestScopeFactory,

		CreateTask:       createTask,
		ConfirmTask:      confirmTask,
		TakeWork:         takeWork,
		FinishWork:       finishWork,
		CheckTask:        checkTask,
		ArchiveTasks:     archiveTasks,
		ExpireProcessing: expireProcessing,
		ResumeDelayed:    resumeDelayed,

		Router: mux,
	}, nil
}
