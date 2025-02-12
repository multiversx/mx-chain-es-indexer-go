package config

// Config will hold the whole config file's data
type Config struct {
	Config struct {
		AvailableIndices []string `toml:"available-indices"`
		ESDTPrefix       string   `toml:"esdt-prefix"`
		AddressConverter struct {
			Length int    `toml:"length"`
			Type   string `toml:"type"`
			Prefix string `toml:"prefix"`
		} `toml:"address-converter"`
		ValidatorKeysConverter struct {
			Length int    `toml:"length"`
			Type   string `toml:"type"`
		} `toml:"validator-keys-converter"`
		Hasher struct {
			Type string `toml:"type"`
		} `toml:"hasher"`
		Marshaller struct {
			Type string `toml:"type"`
		} `toml:"marshaller"`
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
	Sovereign bool
}

// ClusterConfig will hold the config for the Elasticsearch cluster
type ClusterConfig struct {
	Config struct {
		DisabledIndices []string `toml:"disabled-indices"`
		WebSocket       struct {
			URL                string `toml:"url"`
			Mode               string `toml:"mode"`
			DataMarshallerType string `toml:"data-marshaller-type"`
			RetryDurationInSec uint32 `toml:"retry-duration-in-seconds"`
			BlockingAckOnError bool   `toml:"blocking-ack-on-error"`
			WithAcknowledge    bool   `toml:"with-acknowledge"`
			AckTimeoutInSec    uint32 `toml:"acknowledge-timeout-in-seconds"`
		} `toml:"web-socket"`
		ElasticCluster struct {
			UseKibana                 bool   `toml:"use-kibana"`
			URL                       string `toml:"url"`
			UserName                  string `toml:"username"`
			Password                  string `toml:"password"`
			BulkRequestMaxSizeInBytes int    `toml:"bulk-request-max-size-in-bytes"`
		} `toml:"elastic-cluster"`
		MainChainCluster struct {
			Enabled  bool   `toml:"enabled"`
			URL      string `toml:"url"`
			UserName string `toml:"username"`
			Password string `toml:"password"`
		} `toml:"main-chain-elastic-cluster"`
	} `toml:"config"`
}

// ApiRoutesConfig holds the configuration related to Rest API routes
type ApiRoutesConfig struct {
	RestApiInterface string                      `toml:"rest-api-interface"`
	APIPackages      map[string]APIPackageConfig `toml:"api-packages"`
}

// APIPackageConfig holds the configuration for the routes of each package
type APIPackageConfig struct {
	Routes []RouteConfig `toml:"routes"`
}

// RouteConfig holds the configuration for a single route
type RouteConfig struct {
	Name string `toml:"name"`
	Open bool   `toml:"open"`
}
