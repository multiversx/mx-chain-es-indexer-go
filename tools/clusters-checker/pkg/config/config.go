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
		IntervalSettings []struct {
			Start int `toml:"start"`
			Stop  int `toml:"stop"`
		} `toml:"interval"`
		IndicesWithTimestamp []string `toml:"indices-with-timestamp"`
		IndicesNoTimestamp   []string `toml:"indices-no-timestamp"`
	} `toml:"compare"`
}
