package withKibana

// Transactions will hold the configuration for the transactions index
var Transactions = Object{
	"index_patterns": Array{
		"transactions-*",
	},
	"settings": Object{
		"number_of_shards":                                 5,
		"number_of_replicas":                               1,
		"opendistro.index_state_management.policy_id":      "transactions_policy",
		"opendistro.index_state_management.rollover_alias": "transactions",
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
				"type": "date",
			},
		},
	},
}
