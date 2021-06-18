package withKibana

var AccountsESDTHistory = Object{
	"index_patterns": Array{
		"accountsesdthistory-*",
	},
	"settings": Object{
		"number_of_shards":   5,
		"number_of_replicas": 0,
		"opendistro.index_state_management.rollover_alias": "accountsesdthistory",
	},
	"mappings": Object{
		"properties": Object{
			"timestamp": Object{
				"type":   "date",
				"format": "epoch_second",
			},
		},
	},
}
