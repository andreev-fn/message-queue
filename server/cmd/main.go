package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"server/internal/appbuilder"
	"server/internal/config/yamlconfig"
)

const (
	CmdRun              = "run"
	CmdServeAPI         = "serve-api"
	CmdArchiveMessages  = "archive-messages"
	CmdExpireProcessing = "expire-processing"
	CmdResumeDelayed    = "resume-delayed"
)

func main() {
	availableCommands := []string{CmdRun, CmdServeAPI, CmdArchiveMessages, CmdExpireProcessing, CmdResumeDelayed}

	flagSet := flag.NewFlagSet("", flag.ContinueOnError)

	var configPath string
	flagSet.StringVar(&configPath, "c", "config.yaml", "config file path")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	args := flagSet.Args()

	if len(args) < 1 || !slices.Contains(availableCommands, args[0]) {
		fmt.Println("Usage: queue <cmd>")
		fmt.Println("Available commands: " + strings.Join(availableCommands, ", "))
		return
	}

	conf, err := yamlconfig.LoadFromFile(configPath)
	if err != nil {
		log.Fatalf("yamlconfig.LoadFromFile: %v", err)
	}

	app, err := appbuilder.BuildApp(conf, nil)
	if err != nil {
		log.Fatalf("appbuilder.BuildApp: %v", err)
	}

	switch args[0] {
	case CmdRun:
		Run(app)
	case CmdServeAPI:
		ServeAPI(app)
	case CmdArchiveMessages:
		ArchiveMessages(app)
	case CmdExpireProcessing:
		ExpireProcessing(app)
	case CmdResumeDelayed:
		ResumeDelayed(app)
	}
}

func PingDB(db *sql.DB) error {
	pingCtx, closeCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer closeCtx()

	if err := db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("db.PingContext: %w", err)
	}

	return nil
}
