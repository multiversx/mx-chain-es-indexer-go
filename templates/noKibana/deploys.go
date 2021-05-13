package noKibana

// Deploys will hold the configuration for the deploys index
var Deploys = Object{
	"index_patterns": Array{
		"deploys-*",
	},
	"settings": Object{
		"number_of_shards":   3,
		"number_of_replicas": 0,
	},
}
