package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"server/internal/appbuilder"
	"server/internal/config/yamlconfig"
)

const (
	CmdListen           = "listen"
	CmdArchiveMessages  = "archive-messages"
	CmdExpireProcessing = "expire-processing"
	CmdResumeDelayed    = "resume-delayed"
)

func main() {
	availableCommands := []string{CmdListen, CmdArchiveMessages, CmdExpireProcessing, CmdResumeDelayed}

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

	conf, err := yamlconfig.Load(configPath)
	if err != nil {
		log.Fatalf("yamlconfig.Parse: %v", err)
	}

	app, err := appbuilder.BuildApp(conf, nil)
	if err != nil {
		log.Fatalf("appbuilder.BuildApp: %v", err)
	}

	switch args[0] {
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
