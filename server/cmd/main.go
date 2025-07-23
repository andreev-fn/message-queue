package main

import (
	"fmt"
	"log"
	"os"
	"server/cmd/config"
	"server/internal/appbuilder"
	"slices"
	"strings"
)

const (
	CmdListen           = "listen"
	CmdArchiveTasks     = "archive-tasks"
	CmdExpireProcessing = "expire-processing"
	CmdResumeDelayed    = "resume-delayed"
)

func main() {
	availableCommands := []string{CmdListen, CmdArchiveTasks, CmdExpireProcessing, CmdResumeDelayed}

	if len(os.Args) != 2 || !slices.Contains(availableCommands, os.Args[1]) {
		fmt.Println("Usage: queue <cmd>")
		fmt.Println("Available commands: " + strings.Join(availableCommands, ", "))
		return
	}

	conf, err := config.Parse()
	if err != nil {
		log.Fatalf("config.Parse: %v", err)
	}

	app, err := appbuilder.BuildApp(conf, nil)
	if err != nil {
		log.Fatalf("appbuilder.BuildApp: %v", err)
	}

	switch os.Args[1] {
	case CmdListen:
		Listen(app)
	case CmdArchiveTasks:
		ArchiveTasks(app)
	case CmdExpireProcessing:
		ExpireProcessing(app)
	case CmdResumeDelayed:
		ResumeDelayed(app)
	}
}
