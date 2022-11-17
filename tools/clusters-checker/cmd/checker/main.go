package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/tools/clusters-checker/pkg/checkers"
	"github.com/ElrondNetwork/elastic-indexer-go/tools/clusters-checker/pkg/config"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/closing"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go-logger/file"
	"github.com/pelletier/go-toml"
	"github.com/urfave/cli"
)

const configFileName = "config.toml"

const helpTemplate = `NAME:
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

func main() {
	app := cli.NewApp()
	cli.AppHelpTemplate = helpTemplate
	app.Name = "Clusters checker"
	app.Version = "v1.0.0"
	app.Usage = "Clusters checker"
	app.Flags = []cli.Flag{
		configPath, checkCounts, checkNoTimestamp, checkWithTimestamp, checkOnlyIds, logLevel, logSaveFile, disableAnsiColor,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}

	_ = logger.SetLogLevel("*:DEBUG")

	app.Action = checkClusters

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

}

func checkClusters(ctx *cli.Context) {
	cfgPath := ctx.String(configPath.Name)
	cfg, err := loadConfigFile(cfgPath)
	if err != nil {
		log.Error("cannot load config file", "error", err.Error())
		return
	}

	fileLogging, err := initializeLogger(ctx, *cfg)
	if err != nil {
		log.Error("cannot initialize logger", "error", err.Error())
		return
	}

	checkOnlyIDs := ctx.Bool(checkOnlyIds.Name)
	checkCountsFlag := ctx.Bool(checkCounts.Name)
	if checkCountsFlag {
		clusterChecker, errC := checkers.CreateClusterChecker(cfg, &checkers.Interval{}, "instance_0", checkOnlyIDs)
		if errC != nil {
			log.Error("cannot create cluster checker", "error", errC.Error())
			return
		}

		errC = clusterChecker.CompareCounts()
		if errC != nil {
			log.Error("cannot check counts", "error", errC.Error())
			return
		}

		return
	}

	checkIndicesNoTimestampFlag := ctx.Bool(checkNoTimestamp.Name)
	if checkIndicesNoTimestampFlag {
		clusterChecker, errC := checkers.CreateClusterChecker(cfg, &checkers.Interval{}, "instance_0", checkOnlyIDs)
		if errC != nil {
			log.Error("cannot create cluster checker", "error", errC.Error())
			return
		}

		errC = clusterChecker.CompareIndicesNoTimestamp()
		if errC != nil {
			log.Error("cannot check indices", "error", errC.Error())
			return
		}

		return
	}

	checkWithTimestampFlag := ctx.Bool(checkWithTimestamp.Name)
	if checkWithTimestampFlag {
		checkClustersIndexesWithInterval(cfg, checkOnlyIDs)
		return
	}

	if !check.IfNilReflect(fileLogging) {
		err = fileLogging.Close()
		log.LogIfError(err)
	}

	log.Error("no flag has been provided")
}

func checkClustersIndexesWithInterval(cfg *config.Config, checkOnlyIDs bool) {
	wg := sync.WaitGroup{}
	ccs, err := checkers.CreateMultipleCheckers(cfg, checkOnlyIDs)
	if err != nil {
		log.Error("cannot create cluster checker", "error", err.Error())
	}

	wg.Add(len(ccs))
	for _, c := range ccs {
		go func(che checkers.Checker) {
			errC := che.CompareIndicesWithTimestamp()
			if errC != nil {
				log.Error("cannot check indices", "error", errC.Error())
			}
			wg.Done()
		}(c)
	}

	wg.Wait()

}

func loadConfigFile(pathStr string) (*config.Config, error) {
	tomlBytes, err := loadBytesFromFile(path.Join(pathStr, configFileName))
	if err != nil {
		return nil, err
	}

	var cfg config.Config
	err = toml.Unmarshal(tomlBytes, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func loadBytesFromFile(file string) ([]byte, error) {
	return ioutil.ReadFile(file)
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
		DefaultLogsPath: cfg.Logs.LogsPath,
		LogFilePrefix:   cfg.Logs.LogFilePrefix,
	})
	if err != nil {
		return nil, fmt.Errorf("%w creating a log file", err)
	}

	err = fileLogging.ChangeFileLifeSpan(
		time.Second*time.Duration(cfg.Logs.LogFileLifeSpanInSec),
		uint64(cfg.Logs.LogFileLifeSpanInMB),
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
