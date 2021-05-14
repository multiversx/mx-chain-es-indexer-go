package withKibana

// SCDeploys will hold the configuration for the scdeploys index
var SCDeploys = Object{
	"index_patterns": Array{
		"scdeploys-*",
	},
	"settings": Object{
		"number_of_shards":   3,
		"number_of_replicas": 0,
	},
}
