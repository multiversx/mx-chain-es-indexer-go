package withKibana

// Operations will hold the configuration for the operations index
var Operations = Object{
	"index_patterns": Array{
		"operations-*",
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
				"type": "long",
			},
			"timestamp": Object{
				"type":   "date",
				"format": "epoch_second",
			},
		},
	},
}
