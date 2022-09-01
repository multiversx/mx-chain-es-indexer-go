package config

// Config will hold the whole config file's data
type Config struct {
	Config struct {
		EnabledIndices []string `toml:"enabled-indices"`
		ElasticCluster struct {
			UseKibana                 bool   `toml:"use-kibana"`
			URL                       string `toml:"url"`
			UserName                  string `toml:"username"`
			Password                  string `toml:"password"`
			BulkRequestMaxSizeInBytes int    `toml:"bulk-request-max-size-in-bytes"`
		} `toml:"elastic-cluster"`
		WebSocket struct {
			ServerURL           string `toml:"server-url"`
			DataMarshallerTyper string `toml:"data-marshaller-type"`
		} `toml:"web-socket"`
		AddressConverter struct {
			Length int    `toml:"length"`
			Type   string `toml:"type"`
		} `toml:"address-converter"`
		ValidatorKeysConverter struct {
			Length int    `toml:"length"`
			Type   string `toml:"type"`
		} `toml:"validator-keys-converter"`
		Hasher struct {
			Type string `toml:"type"`
		} `toml:"hasher"`
		Economics struct {
			Denomination int `toml:"denomination"`
		} `toml:"economics"`
		Logs struct {
			LogFileLifeSpanInMB  int    `toml:"log-file-life-span-in-mb"`
			LogFileLifeSpanInSec int    `toml:"log-file-life-span-in-sec"`
			LogFilePrefix        string `toml:"log-file-prefix"`
			LogsPath             string `toml:"logs-path"`
		} `toml:"logs"`
	} `toml:"config"`
}