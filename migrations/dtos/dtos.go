package dtos

import "github.com/multiversx/mx-chain-es-indexer-go/data"

type ClusterSettings struct {
	URL      string
	User     string
	Password string
}

type MigrationInfo struct {
	Status    string `json:"status"`
	Timestamp uint64 `json:"timestamp"`
}

type ResponseLogsSearch struct {
	Hits struct {
		Hits []struct {
			ID     string     `json:"_id"`
			Source *data.Logs `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}
