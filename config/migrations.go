package config

type MigrationsConfig struct {
	SourceCluster struct {
		URL      string `toml:"url"`
		UserName string `toml:"username"`
		Password string `toml:"password"`
	} `toml:"source-cluster"`
	Migrations []struct {
		Start            string `toml:"start"`
		Name             string `toml:"name"`
		Description      string `toml:"description"`
		SourceIndex      string `toml:"source-index"`
		DestinationIndex string `toml:"destination-index"`
	} `toml:"migrations"`
}
