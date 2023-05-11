package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/closing"
	"github.com/multiversx/mx-chain-es-indexer-go/config"
	"github.com/multiversx/mx-chain-es-indexer-go/factory"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-logger-go/file"
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
		configurationPreferencesFile,
		logLevel,
		logSaveFile,
		disableAnsiColor,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The MultiversX Team",
			Email: "contact@multiversx.com",
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
	cfg, err := loadMainConfig(ctx.GlobalString(configurationFile.Name))
	if err != nil {
		return err
	}

	clusterCfg, err := loadClusterConfig(ctx.GlobalString(configurationPreferencesFile.Name))
	if err != nil {
		return err
	}

	fileLogging, err := initializeLogger(ctx, cfg)
	if err != nil {
		return err
	}

	wsHost, err := factory.CreateWsIndexer(cfg, clusterCfg)
	if err != nil {
		log.Error("cannot create ws indexer", "error", err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-interrupt
	log.Info("closing app at user's signal")
	err = wsHost.Close()
	if err != nil {
		log.Error("cannot close ws indexer", "error", err)
	}

	if !check.IfNilReflect(fileLogging) {
		err = fileLogging.Close()
		log.LogIfError(err)
	}
	return nil
}

func loadMainConfig(filepath string) (config.Config, error) {
	cfg := config.Config{}
	err := core.LoadTomlFile(&cfg, filepath)

	return cfg, err
}

func loadClusterConfig(filepath string) (config.ClusterConfig, error) {
	cfg := config.ClusterConfig{}
	err := core.LoadTomlFile(&cfg, filepath)

	return cfg, err
}

func initializeLogger(ctx *cli.Context, cfg config.Config) (closing.Closer, error) {
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

	disableAnsi := ctx.GlobalBool(disableAnsiColor.Name)
	err = removeANSIColorsForLoggerIfNeeded(disableAnsi)
	if err != nil {
		return nil, err
	}

	return fileLogging, nil
}

func removeANSIColorsForLoggerIfNeeded(disableAnsi bool) error {
	if !disableAnsi {
		return nil
	}

	err := logger.RemoveLogObserver(os.Stdout)
	if err != nil {
		return err
	}

	return logger.AddLogObserver(os.Stdout, &logger.PlainFormatter{})
}
