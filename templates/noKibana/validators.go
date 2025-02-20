package noKibana

// Validators will hold the configuration for the validators index
var Validators = Object{
	"index_patterns": Array{
		"validators-*",
	},
	"template": Object{
		"settings": Object{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": Object{
			"properties": Object{
				"publicKeys": Object{
					"type": "keyword",
				},
			},
		},
	},
}
