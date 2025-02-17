package noKibana

import (
	"github.com/multiversx/mx-chain-es-indexer-go/templates"
)

type Array = templates.Array
type Object = templates.Object

// Accounts will hold the configuration for the accounts index
var Accounts = Object{
	"index_patterns": Array{
		"accounts-*",
	},
	"template": Object{
		"settings": Object{
			"number_of_shards":   3,
			"number_of_replicas": 0,
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
					"type": "double",
				},
				"address": Object{
					"type": "keyword",
				},
				"balance": Object{
					"type": "keyword",
				},
				"totalBalanceWithStake": Object{
					"type": "keyword",
				},
				"shardID": Object{
					"type": "long",
				},
				"timestamp": Object{
					"type":   "date",
					"format": "epoch_second",
				},
				"userName": Object{
					"type": "keyword",
				},
				"owner": Object{
					"type": "keyword",
				},
				"developerRewards": Object{
					"type": "keyword",
				},
				"developerRewardsNum": Object{
					"type": "double",
				},
			},
		},
	},
}
