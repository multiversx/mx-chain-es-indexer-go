package noKibana

// Events will hold the configuration for the events index
var Events = Object{
	"index_patterns": Array{
		"events-*",
	},
	"template": Object{
		"settings": Object{
			"number_of_shards":   5,
			"number_of_replicas": 0,
		},
		"mappings": Object{
			"properties": Object{
				"txHash": Object{
					"type": "keyword",
				},
				"originalTxHash": Object{
					"type": "keyword",
				},
				"logAddress": Object{
					"type": "keyword",
				},
				"address": Object{
					"type": "keyword",
				},
				"identifier": Object{
					"type": "keyword",
				},
				"shardID": Object{
					"type": "long",
				},
				"data": Object{
					"index": "false",
					"type":  "text",
				},
				"additionalData": Object{
					"type": "text",
				},
				"topics": Object{
					"type": "text",
				},
				"order": Object{
					"type": "long",
				},
				"txOrder": Object{
					"type": "long",
				},
				"timestamp": Object{
					"type":   "date",
					"format": "epoch_second",
				},
			},
		},
	},
}
