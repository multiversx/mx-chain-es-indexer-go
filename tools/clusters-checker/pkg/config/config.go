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
		Indices []string `toml:"indices"`
	} `toml:"compare"`
}
