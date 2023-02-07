package noKibana

// AccountsHistory will hold the configuration for the accountshistory index
var AccountsHistory = Object{
	"index_patterns": Array{
		"accountshistory-*",
	},
	"settings": Object{
		"number_of_shards":   5,
		"number_of_replicas": 0,
	},
	"mappings": Object{
		"properties": Object{
			"address": Object{
				"type": "keyword",
			},
			"balance": Object{
				"type": "keyword",
			},
			"isSender": Object{
				"type": "boolean",
			},
			"isSmartContract": Object{
				"type": "boolean",
			},
			"shardID": Object{
				"type": "long",
			},
			"timestamp": Object{
				"type":   "date",
				"format": "epoch_second",
			},
		},
	},
}
