package main

import (
	"github.com/ElrondNetwork/elastic-indexer-go/tools/clusters-checker/pkg/checkers"
	"io/ioutil"
	"os"
	"path"

	"github.com/ElrondNetwork/elastic-indexer-go/tools/clusters-checker/pkg/config"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/pelletier/go-toml"
	"github.com/urfave/cli"
)

const configFileName = "config.toml"

var (
	log = logger.GetOrCreate("main")

	// defines the path to the config folder
	configPath = cli.StringFlag{
		Name:  "config-path",
		Usage: "The path to the config folder",
		Value: "./",
	}
)

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
	app.Name = "Cluster checker"
	app.Version = "v1.0.0"
	app.Usage = "Cluster checker"
	app.Flags = []cli.Flag{
		configPath,
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
	}

	clusterChecker, err := checkers.CreateClusterChecker(cfg)
	if err != nil {
		log.Error("cannot create cluster checker", "error", err.Error())
	}

	err = clusterChecker.CompareCounts()
	if err != nil {
		log.Error("cannot check counts", "error", err.Error())
	}

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
