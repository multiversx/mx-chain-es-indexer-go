package withKibana

// Validators will hold the configuration for the validators index
var Validators = Object{
	"index_patterns": Array{
		"validators-*",
	},
	"settings": Object{
		"number_of_shards":                                 1,
		"number_of_replicas":                               0,
		"opendistro.index_state_management.policy_id":      "validators_policy",
		"opendistro.index_state_management.rollover_alias": "validators",
	},
}
