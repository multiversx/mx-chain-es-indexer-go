package noKibana

// Delegators will hold the configuration for the delegators index
var Delegators = Object{
	"index_patterns": Array{
		"delegators-*",
	},
	"settings": Object{
		"number_of_shards":   3,
		"number_of_replicas": 0,
	},

	"mappings": Object{
		"properties": Object{
			"activeStakeNum": Object{
				"type": "double",
			},
		},
	},
}
