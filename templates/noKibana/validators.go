package noKibana

// Validators will hold the configuration for the validators index
var Validators = Object{
	"index_patterns": Array{
		"validators-*",
	},
	"settings": Object{
		"number_of_shards":   1,
		"number_of_replicas": 0,
	},
}
