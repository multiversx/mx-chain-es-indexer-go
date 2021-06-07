package noKibana

// Tokens will hold the configuration for the tokens index
var Tokens = Object{
	"index_patterns": Array{
		"tokens-*",
	},
	"settings": Object{
		"number_of_shards":   3,
		"number_of_replicas": 0,
	},
}
