package main

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/urfave/cli"
)

var (
	configurationFile = cli.StringFlag{
		Name:  "config",
		Usage: "The main configuration file to load",
		Value: "./config/config.toml",
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
)
