package noKibana

// Values will hold the configuration for the values index
var Values = Object{
	"index_patterns": Array{
		"values-*",
	},
	"settings": Object{
		"number_of_shards":   1,
		"number_of_replicas": 0,
	},

	"mappings": Object{
		"properties": Object{
			"key": Object{
				"type": "keyword",
			},
			"value": Object{
				"type": "keyword",
			},
		},
	},
}
