package config

type MigrationsConfig struct {
	SourceCluster ClusterInfo  `toml:"source-cluster"`
	Migrations    []*Migration `toml:"migrations"`
}

type Migration struct {
	Start            string `toml:"start"`
	Name             string `toml:"name"`
	Description      string `toml:"description"`
	SourceIndex      string `toml:"source-index"`
	DestinationIndex string `toml:"destination-index"`
}

type ClusterInfo struct {
	URL      string `toml:"url"`
	UserName string `toml:"username"`
	Password string `toml:"password"`
}
