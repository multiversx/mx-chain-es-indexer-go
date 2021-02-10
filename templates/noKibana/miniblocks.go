package noKibana

// Miniblocks will hold the configuration for the miniblocks index
var Miniblocks = Object{
	"index_patterns": Array{
		"miniblocks-*",
	},
	"settings": Object{
		"number_of_shards":   3,
		"number_of_replicas": 0,
	},
}
