package main

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/urfave/cli"
)

var (
	log = logger.GetOrCreate("main")

	// defines the path to the config folder
	configPath = cli.StringFlag{
		Name:  "config-path",
		Usage: "The path to the config folder",
		Value: "./",
	}
	checkCounts = cli.BoolFlag{
		Name:  "check-counts",
		Usage: "If set, the checker will verify the counts between clusters",
	}
	checkWithTimestamp = cli.BoolFlag{
		Name:  "check-with-timestamp",
		Usage: "If set, the checker will verify all the indices from list with timestamp",
	}
	checkNoTimestamp = cli.BoolFlag{
		Name:  "check-no-timestamp",
		Usage: "If set, the checker will verify the indices from list with no timestamp",
	}
	checkOnlyIds = cli.BoolFlag{
		Name:  "check-only-ids",
		Usage: "If set, the checker will verify only the ids",
	}
	logLevel = cli.StringFlag{
		Name: "log-level",
		Usage: "This flag specifies the logger `level(s)`. It can contain multiple comma-separated value. For example" +
			", if set to *:INFO the logs for all packages will have the INFO level. However, if set to *:INFO,api:DEBUG" +
			" the logs for all packages will have the INFO level, excepting the api package which will receive a DEBUG" +
			" log level.",
		Value: "*:" + logger.LogInfo.String(),
	}
	logSaveFile = cli.BoolFlag{
		Name:  "log-save",
		Usage: "Boolean option for enabling log saving. If set, it will automatically save all the logs into a file.",
	}
	// enableAnsiColor defines if the logger subsystem should displaying ANSI colors
	enableAnsiColor = cli.BoolFlag{
		Name:  "enable-ansi-color",
		Usage: "Boolean option for enabling ANSI colors in the logging system.",
	}
)
