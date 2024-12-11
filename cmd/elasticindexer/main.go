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
	"github.com/multiversx/mx-chain-core-go/data/outport"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-logger-go/file"
	"github.com/urfave/cli"

	"github.com/multiversx/mx-chain-es-indexer-go/config"
	"github.com/multiversx/mx-chain-es-indexer-go/factory"
	"github.com/multiversx/mx-chain-es-indexer-go/metrics"
	"github.com/multiversx/mx-chain-es-indexer-go/process/wsindexer"
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

// appVersion should be populated at build time using ldflags
// Usage examples:
// linux/mac:
//
//	go build -v -ldflags="-X main.appVersion=$(git describe --tags --long --dirty)"
//
// windows:
//
//	for /f %i in ('git describe --tags --long --dirty') do set VERS=%i
//	go build -v -ldflags="-X main.version=%VERS%"
var version = "undefined"

func main() {
	app := cli.NewApp()
	cli.AppHelpTemplate = helpTemplate
	app.Name = "Elastic indexer"
	app.Usage = "This tool will index data in an Elasticsearch database"
	app.Flags = []cli.Flag{
		configurationFile,
		configurationPreferencesFile,
		configurationApiFile,
		logLevel,
		logSaveFile,
		disableAnsiColor,
		sovereign,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The MultiversX Team",
			Email: "contact@multiversx.com",
		},
	}

	app.Version = version
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
		return fmt.Errorf("%w while loading the config file", err)
	}
	cfg.Sovereign = ctx.GlobalBool(sovereign.Name)

	clusterCfg, err := loadClusterConfig(ctx.GlobalString(configurationPreferencesFile.Name))
	if err != nil {
		return fmt.Errorf("%w while loading the preferences config file", err)
	}

	fileLogging, err := initializeLogger(ctx, cfg)
	if err != nil {
		return fmt.Errorf("%w while initializing the logger", err)
	}

	statusMetrics := metrics.NewStatusMetrics()
	wsHost, err := factory.CreateWsIndexer(cfg, clusterCfg, statusMetrics, ctx.App.Version)
	if err != nil {
		return fmt.Errorf("%w while creating the indexer", err)
	}

	apiConfig, err := loadApiConfig(ctx.GlobalString(configurationApiFile.Name))
	if err != nil {
		return fmt.Errorf("%w while loading the api config file", err)
	}

	webServer, err := factory.CreateWebServer(apiConfig, statusMetrics)
	if err != nil {
		return fmt.Errorf("%w while creating the web server", err)
	}

	err = webServer.StartHttpServer()
	if err != nil {
		return fmt.Errorf("%w while starting the web server", err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	retryDuration := time.Duration(clusterCfg.Config.WebSocket.RetryDurationInSec) * time.Second
	closed := requestSettings(wsHost, retryDuration, interrupt)
	if !closed {
		<-interrupt
	}

	log.Info("closing app at user's signal")
	err = wsHost.Close()
	if err != nil {
		log.Error("cannot close ws indexer", "error", err)
	}

	err = webServer.Close()
	if err != nil {
		log.Error("cannot close web server", "error", err)
	}

	if !check.IfNilReflect(fileLogging) {
		err = fileLogging.Close()
		log.LogIfError(err)
	}
	return nil
}

func requestSettings(host wsindexer.WSClient, retryDuration time.Duration, close chan os.Signal) bool {
	timer := time.NewTimer(0)
	defer timer.Stop()

	emptyMessage := make([]byte, 0)
	for {
		select {
		case <-timer.C:
			err := host.Send(emptyMessage, outport.TopicSettings)
			if err == nil {
				return false
			}
			log.Debug("unable to request settings - will retry", "error", err)

			timer.Reset(retryDuration)
		case <-close:
			return true
		}
	}
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

// loadApiConfig returns a ApiRoutesConfig by reading the config file provided
func loadApiConfig(filepath string) (config.ApiRoutesConfig, error) {
	cfg := config.ApiRoutesConfig{}
	err := core.LoadTomlFile(&cfg, filepath)
	if err != nil {
		return config.ApiRoutesConfig{}, err
	}

	return cfg, nil
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
