package noKibana

// Tags will hold the configuration for the tags index
var Tags = Object{
	"index_patterns": Array{
		"tags-*",
	},
	"template": Object{
		"settings": Object{
			"number_of_shards":   3,
			"number_of_replicas": 0,
		},
		"mappings": Object{
			"properties": Object{
				"count": Object{
					"type": "long",
				},
				"tag": Object{
					"type": "keyword",
				},
			},
		},
	},
}
