package noKibana

// Rounds will hold the configuration for the rounds index
var Rounds = Object{
	"index_patterns": Array{
		"rounds-*",
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
				"blockWasProposed": Object{
					"type": "boolean",
				},
				"epoch": Object{
					"type": "long",
				},
				"round": Object{
					"type": "double",
				},
				"shardId": Object{
					"type": "long",
				},
				"signersIndexes": Object{
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
