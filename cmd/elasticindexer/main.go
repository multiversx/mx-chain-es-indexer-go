package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/config"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go-logger/file"
	"github.com/urfave/cli"
)

var (
	log          = logger.GetOrCreate("indexer")
	helpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
VERSION:
   {{.Version}}
   {{end}}
`
)

func main() {
	app := cli.NewApp()
	cli.AppHelpTemplate = helpTemplate
	app.Name = "Elastic indexer"
	app.Usage = "This tool will index data in an Elasticsearch database"
	app.Flags = []cli.Flag{
		configurationFile,
		logLevel,
		logSaveFile,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}

	app.Action = startIndexer

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startIndexer(ctx *cli.Context) error {
	configurationFileName := ctx.GlobalString(configurationFile.Name)
	cfg, err := loadMainConfig(configurationFileName)
	if err != nil {
		return err
	}

	fileLogging, err := initializeLogger(ctx, cfg)
	if err != nil {
		return err
	}

	log.Info("closing indexer")
	if !check.IfNil(fileLogging) {
		err = fileLogging.Close()
		log.LogIfError(err)
	}

	return nil
}

func loadMainConfig(filepath string) (*config.Config, error) {
	cfg := &config.Config{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func initializeLogger(ctx *cli.Context, cfg *config.Config) (file.FileLoggingHandler, error) {
	logLevelFlagValue := ctx.GlobalString(logLevel.Name)
	err := logger.SetLogLevel(logLevelFlagValue)
	if err != nil {
		return nil, err
	}

	withLogFile := ctx.GlobalBool(logSaveFile.Name)
	if !withLogFile {
		return nil, nil
	}

	workingDir, err := os.Getwd()
	if err != nil {
		log.LogIfError(err)
		workingDir = ""
	}

	fileLogging, err := file.NewFileLogging(file.ArgsFileLogging{
		WorkingDir:      workingDir,
		DefaultLogsPath: cfg.Config.Logs.LogsPath,
		LogFilePrefix:   cfg.Config.Logs.LogFilePrefix,
	})
	if err != nil {
		return nil, fmt.Errorf("%w creating a log file", err)
	}

	err = fileLogging.ChangeFileLifeSpan(
		time.Second*time.Duration(cfg.Config.Logs.LogFileLifeSpanInSec),
		uint64(cfg.Config.Logs.LogFileLifeSpanInMB),
	)
	if err != nil {
		return nil, err
	}

	return fileLogging, nil
}
