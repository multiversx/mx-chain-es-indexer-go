package noKibana

// Miniblocks will hold the configuration for the miniblocks index
var Miniblocks = Object{
	"index_patterns": Array{
		"miniblocks-*",
	},
	"template": Object{
		"settings": Object{
			"number_of_shards":   3,
			"number_of_replicas": 0,
		},
		"mappings": Object{
			"properties": Object{
				"procTypeD": Object{
					"type": "keyword",
				},
				"procTypeS": Object{
					"type": "keyword",
				},
				"receiverBlockHash": Object{
					"type": "keyword",
				},
				"receiverShard": Object{
					"type": "long",
				},
				"reserved": Object{
					"index": "false",
					"type":  "keyword",
				},
				"senderBlockHash": Object{
					"type": "keyword",
				},
				"senderShard": Object{
					"type": "long",
				},
				"timestamp": Object{
					"type":   "date",
					"format": "epoch_second",
				},
				"type": Object{
					"type": "keyword",
				},
			},
		},
	},
}
