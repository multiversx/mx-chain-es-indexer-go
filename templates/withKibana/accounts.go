package withKibana

import (
	"github.com/ElrondNetwork/elastic-indexer-go/templates"
)

type Object = templates.Object
type Array = templates.Array

// Accounts will hold the configuration for the accounts index
var Accounts = Object{
	"index_patterns": Array{
		"accounts-*",
	},
	"settings": Object{
		"number_of_shards":   3,
		"number_of_replicas": 0,
	},
	"mappings": Object{
		"properties": Object{
			"balanceNum": Object{
				"type": "double",
			},
		},
	},
}
