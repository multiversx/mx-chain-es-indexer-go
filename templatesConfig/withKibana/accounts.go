package withKibana

import (
	"github.com/ElrondNetwork/elastic-indexer-go/templatesConfig"
)

type Object = templatesConfig.Object
type Array = templatesConfig.Array

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
