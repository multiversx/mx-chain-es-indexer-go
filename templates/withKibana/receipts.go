package withKibana

// Receipts will hold the configuration for the receipts index
var Receipts = Object{
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
		"opendistro.index_state_management.rollover_alias": "receipts",
	},
	"mappings": Object{
		"properties": Object{
			"timestamp": Object{
				"type": "date",
			},
		},
	},
}
