package withKibana

var TPS = Object{
	"index_patterns": Array{
		"tps-*",
	},
	"settings": Object{
		"number_of_shards":   3,
		"number_of_replicas": 0,
	},
}
