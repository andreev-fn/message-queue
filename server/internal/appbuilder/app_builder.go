package appbuilder

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"server/internal/routes"
	"server/internal/storage"
	"server/internal/usecases"
	"server/internal/utils/timeutils"

	_ "github.com/jackc/pgx/v4/stdlib"
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

	archiveTasks := usecases.NewArchiveTasks(clock, db, taskRepo, archivedTaskRepo)
	expireProcessing := usecases.NewExpireProcessing(clock, logger, db, taskRepo)
	resumeDelayed := usecases.NewResumeDelayed(clock, logger, db, taskRepo)

	mux := http.NewServeMux()
	routes.NewCreateTask(
		db,
		logger,
		usecases.NewCreateTask(logger, clock, db, taskRepo),
	).Mount(mux)
	routes.NewFinishWork(
		db,
		logger,
		usecases.NewFinishWork(clock, logger, db, taskRepo),
	).Mount(mux)
	routes.NewTakeWork(
		db,
		logger,
		usecases.NewTakeWork(logger, clock, db, taskRepo),
	).Mount(mux)
	routes.NewConfirmTask(
		db,
		logger,
		usecases.NewConfirmTask(logger, clock, db, taskRepo),
	).Mount(mux)
	routes.NewCheckTask(
		db,
		logger,
		usecases.NewCheckTask(db, taskRepo, archivedTaskRepo),
	).Mount(mux)

	return &App{
		clock,
		logger,
		db,
		taskRepo,
		archivedTaskRepo,
		archiveTasks,
		expireProcessing,
		resumeDelayed,
		mux,
	}, nil
}
