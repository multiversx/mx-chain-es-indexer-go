package config

type Config struct {
	SourceCluster struct {
		URL      string `toml:"url"`
		User     string `toml:"user"`
		Password string `toml:"password"`
	} `toml:"source-cluster"`
	DestinationCluster struct {
		URL      string `toml:"url"`
		User     string `toml:"user"`
		Password string `toml:"password"`
	} `toml:"destination-cluster"`
	Compare struct {
		BlockchainStartTime  int64    `toml:"blockchain-start-time"`
		NumParallelReads     int      `toml:"num-parallel-reads"`
		IndicesWithTimestamp []string `toml:"indices-with-timestamp"`
		IndicesNoTimestamp   []string `toml:"indices-no-timestamp"`
	} `toml:"compare"`
	Logs struct {
		LogFileLifeSpanInMB  int    `toml:"log-file-life-span-in-mb"`
		LogFileLifeSpanInSec int    `toml:"log-file-life-span-in-sec"`
		LogFilePrefix        string `toml:"log-file-prefix"`
		LogsPath             string `toml:"logs-path"`
	} `toml:"logs"`
}
