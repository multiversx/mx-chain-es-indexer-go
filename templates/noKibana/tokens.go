package noKibana

// Tokens will hold the configuration for the tokens index
var Tokens = Object{
	"index_patterns": Array{
		"tokens-*",
	},
	"settings": Object{
		"number_of_shards":   3,
		"number_of_replicas": 0,
	},
	"mappings": Object{
		"properties": Object{
			"timestamp": Object{
				"type": "date",
			},
			"metaData": Object{
				"type": "nested",
				"properties": Object{
					"name": Object{
						"type": "text",
					},
					"creator": Object{
						"type": "text",
					},
					"attributes": Object{
						"type": "nested",
						"properties": Object{
							"tags": Object{
								"type": "text",
							},
						},
					},
				},
			},
		},
	},
}
