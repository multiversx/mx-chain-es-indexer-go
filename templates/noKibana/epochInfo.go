package noKibana

// EpochInfo will hold the configuration for the epochinfo index
var EpochInfo = Object{
	"index_patterns": Array{
		"epochinfo-*",
	},
	"settings": Object{
		"number_of_shards":   3,
		"number_of_replicas": 0,
	},
}
