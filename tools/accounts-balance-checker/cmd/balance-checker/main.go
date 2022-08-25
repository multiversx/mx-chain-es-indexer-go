package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ElrondNetwork/elastic-indexer-go/tools/accounts-balance-checker/pkg/check"
	"github.com/ElrondNetwork/elastic-indexer-go/tools/accounts-balance-checker/pkg/config"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/urfave/cli"
)

var (
	log = logger.GetOrCreate("main")

	configFile = cli.StringFlag{
		Name:  "config-file",
		Value: "config.json",
	}
	checkBalanceEGLD = cli.BoolFlag{
		Name:  "check-balance-egld",
		Usage: "If set, the checker will verify all the balance value of the accounts with EGLD",
	}
	checkBalanceESDT = cli.BoolFlag{
		Name:  "check-balance-esdt",
		Usage: "If set, the checker wil verify all the balance value of the accounts with ESDT",
	}
	repairFlag = cli.BoolFlag{
		Name:  "repair",
		Usage: "If set, the checker will also repair the wrong balances",
	}

	logLevel = cli.StringFlag{
		Name:  "log-level",
		Value: "*:INFO",
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "Elasticsearch accounts balance checker"
	app.Version = "v1.0.0"
	app.Usage = "This is the entry point for Elasticsearch accounts balance checker tool"
	app.Flags = []cli.Flag{
		configFile,
		checkBalanceEGLD,
		checkBalanceESDT,
		logLevel,
		repairFlag,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}

	app.Action = startCheck
	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startCheck(ctx *cli.Context) {
	setLogLevelDebug(ctx)

	cfg, err := readConfig(ctx)
	if err != nil {
		log.Error("cannot read config file", "error", err)
		return
	}

	repair := ctx.Bool(repairFlag.Name)

	balanceChecker, err := check.CreateBalanceChecker(cfg, repair)
	if err != nil {
		log.Error("cannot create balance checker", "error", err)
		return
	}

	shouldCheckBalanceEGLD := ctx.Bool(checkBalanceEGLD.Name)
	if shouldCheckBalanceEGLD {
		err = balanceChecker.CheckEGLDBalances()
		if err != nil {
			log.Error("cannot check balance EGLD", "error", err)
			return
		}

		log.Info("done")
	}

	shouldCheckBalanceESDT := ctx.Bool(checkBalanceESDT.Name)
	if shouldCheckBalanceESDT {
		err = balanceChecker.CheckESDTBalances()
		if err != nil {
			log.Error("cannot check balance ESDT", "error", err)
			return
		}
	}

	if !shouldCheckBalanceEGLD && !shouldCheckBalanceESDT {
		log.Error("no flag has been provided")
	}

	return
}

func readConfig(ctx *cli.Context) (*config.Config, error) {
	jsonFile, err := ioutil.ReadFile(ctx.String(configFile.Name))
	if err != nil {
		return nil, err
	}
	cfg := &config.Config{}
	err = json.Unmarshal(jsonFile, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func setLogLevelDebug(ctx *cli.Context) {
	err := logger.SetLogLevel(ctx.String(logLevel.Name))
	if err != nil {
		fmt.Println("main logger.SetLogLevel error: ", err.Error())
	}

	err = logger.RemoveLogObserver(os.Stdout)
	if err != nil {
		fmt.Println("main logger.RemoveLogObserver error: ", err.Error())
	}

	err = logger.AddLogObserver(os.Stdout, &logger.PlainFormatter{})
	if err != nil {
		fmt.Println("main logger.AddLogObserver error: ", err.Error())
	}
}
