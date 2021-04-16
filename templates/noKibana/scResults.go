package noKibana

// SCResults will hold the configuration for the scresults index
var SCResults = Object{
	"index_patterns": Array{
		"receipts-*",
	},
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
			"timestamp": Object{
				"type": "date",
			},
		},
	},
}
