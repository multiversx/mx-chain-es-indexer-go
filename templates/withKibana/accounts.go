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
		"opendistro.index_state_management.rollover_alias": "accounts",
	},
	"mappings": Object{
		"properties": Object{
			"balanceNum": Object{
				"type": "double",
			},
			"totalBalanceWithStakeNum": Object{
				"type": "double",
			},
			"nonce": Object{
				"type": "unsigned_long",
			},
			"address": Object{
				"type": "text",
			},
			"balance": Object{
				"type": "text",
			},
			"totalBalanceWithStake": Object{
				"type": "text",
			},
		},
	},
}
