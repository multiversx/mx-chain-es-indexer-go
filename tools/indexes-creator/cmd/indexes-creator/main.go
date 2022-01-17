package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ElrondNetwork/elastic-indexer-go/client"
	"github.com/ElrondNetwork/elastic-indexer-go/client/logging"
	"github.com/ElrondNetwork/elastic-indexer-go/tools/indexes-creator/reader"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/pelletier/go-toml"
	"github.com/urfave/cli"
)

const configFileName = "cluster.toml"

type Config struct {
	ClusterConfig struct {
		URL            string   `toml:"url"`
		Username       string   `toml:"username"`
		Password       string   `toml:"password"`
		UseKibana      bool     `toml:"use-kibana"`
		EnabledIndexes []string `toml:"enabled-indexes"`
	} `toml:"config"`
}

var (
	log = logger.GetOrCreate("main")

	// defines the path to the config folder
	configPath = cli.StringFlag{
		Name:  "config-path",
		Usage: "The path to the config folder",
		Value: "./config/",
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
	app.Name = "Index cr"
	app.Version = "v1.0.0"
	app.Usage = "This is the entry point for Elasticsearch reindexing tool"
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

	app.Action = createIndexesAndMappings

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

}

func createIndexesAndMappings(ctx *cli.Context) {
	cfgPath := ctx.String(configPath.Name)
	cfg, err := loadConfigFile(cfgPath)
	if err != nil {
		log.Error("cannot load config file", "error", err.Error())
	}

	pathToMappings := cfgPath + "noKibana"
	if cfg.ClusterConfig.UseKibana {
		pathToMappings = cfgPath + "withKibana"
	}

	indexesMappings, _, err := reader.GetElasticTemplatesAndPolicies(pathToMappings, cfg.ClusterConfig.EnabledIndexes)
	if err != nil {
		log.Error("cannot load templates", "error", err.Error())
	}

	err = createTemplates(cfg, indexesMappings)
	if err != nil {
		log.Error("cannot create templates", "error", err.Error())
	}

	log.Info("All indexes were created")
}

func createTemplates(cfg *Config, indexesMappings map[string]*bytes.Buffer) error {
	databaseClient, err := client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{cfg.ClusterConfig.URL},
		Username:  cfg.ClusterConfig.Username,
		Password:  cfg.ClusterConfig.Password,
		Logger:    &logging.CustomLogger{},
	})
	if err != nil {
		return err
	}

	for index, indexData := range indexesMappings {
		errCheck := databaseClient.CheckAndCreateTemplate(index, indexData)
		if errCheck != nil {
			return fmt.Errorf("index: %s, error: %w", index, errCheck)
		}

		indexName := fmt.Sprintf("%s-%s", index, "000001")
		errCreate := databaseClient.CheckAndCreateIndex(indexName)
		if errCreate != nil {
			return fmt.Errorf("index: %s, error: %w", index, errCreate)
		}

		errAlias := databaseClient.CheckAndCreateAlias(index, indexName)
		if err != nil {
			return errAlias
		}
	}

	return nil
}

func loadConfigFile(path string) (*Config, error) {
	tomlBytes, err := loadBytesFromFile(path + configFileName)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = toml.Unmarshal(tomlBytes, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func loadBytesFromFile(file string) ([]byte, error) {
	return ioutil.ReadFile(file)
}
