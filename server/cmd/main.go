package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"server/cmd/config"
	"server/internal/appbuilder"
)

const (
	CmdListen           = "listen"
	CmdArchiveMessages  = "archive-messages"
	CmdExpireProcessing = "expire-processing"
	CmdResumeDelayed    = "resume-delayed"
)

func main() {
	availableCommands := []string{CmdListen, CmdArchiveMessages, CmdExpireProcessing, CmdResumeDelayed}

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
	case CmdArchiveMessages:
		ArchiveMessages(app)
	case CmdExpireProcessing:
		ExpireProcessing(app)
	case CmdResumeDelayed:
		ResumeDelayed(app)
	}
}
