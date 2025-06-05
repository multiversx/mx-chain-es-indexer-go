package noKibana

// Rating will hold the configuration for the rating index
var Rating = Object{
	"index_patterns": Array{
		"rating-*",
	},
	"template": Object{
		"settings": Object{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": Object{
			"properties": Object{
				"rating": Object{
					"type": "double",
				},
			},
		},
	},
}
