package noKibana

// Collections will hold the configuration for the collections index
var Collections = Object{
	"index_patterns": Array{
		"collections-*",
	},
	"settings": Object{
		"number_of_shards":   3,
		"number_of_replicas": 0,
	},

	"mappings": Object{
		"dynamic": false,
	},
}
