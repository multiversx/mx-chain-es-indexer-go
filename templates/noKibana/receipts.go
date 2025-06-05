package noKibana

// Receipts will hold the configuration for the receipts index
var Receipts = Object{
	"index_patterns": Array{
		"receipts-*",
	},
	"template": Object{
		"settings": Object{
			"number_of_shards":   3,
			"number_of_replicas": 0,
			"index": Object{
				"sort.field": Array{
					"timestamp",
				},
				"sort.order": Array{
					"desc",
				},
			},
		},
		"mappings": Object{
			"properties": Object{
				"data": Object{
					"type": "keyword",
				},
				"sender": Object{
					"type": "keyword",
				},
				"timestamp": Object{
					"type":   "date",
					"format": "epoch_second",
				},
				"txHash": Object{
					"type": "keyword",
				},
				"value": Object{
					"index": "false",
					"type":  "keyword",
				},
			},
		},
	},
}
