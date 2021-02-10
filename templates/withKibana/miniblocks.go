package withKibana

// Miniblocks will hold the configuration for the miniblocks index
var Miniblocks = Object{
	"index_patterns": Array{
		"miniblocks-*",
	},
	"settings": Object{
		"number_of_shards":                                 3,
		"number_of_replicas":                               0,
		"opendistro.index_state_management.policy_id":      "miniblocks_policy",
		"opendistro.index_state_management.rollover_alias": "miniblocks",
	},
}
