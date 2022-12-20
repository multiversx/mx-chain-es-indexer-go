package config

type Config struct {
	Elasticsearch struct {
		URL      string `json:"url"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	Proxy struct {
		URL                         string `json:"url"`
		MaxNumberOfParallelRequests int    `json:"parallel-requests"`
	} `json:"proxy"`
}
