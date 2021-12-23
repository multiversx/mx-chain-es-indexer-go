package noKibana

// Transactions will hold the configuration for the transactions index
var Transactions = Object{
	"index_patterns": Array{
		"transactions-*",
	},
	"settings": Object{
		"number_of_shards":   5,
		"number_of_replicas": 0,
		"index": Object{
			"sort.field": Array{
				"timestamp", "nonce",
			},
			"sort.order": Array{
				"desc", "desc",
			},
		},
	},

	"mappings": Object{
		"properties": Object{
			"nonce": Object{
				"type": "double",
			},
			"timestamp": Object{
				"type":   "date",
				"format": "epoch_second",
			},
			"gasLimit": Object{
				"type": "double",
			},
			"gasPrice": Object{
				"type": "double",
			},
		},
	},
}
