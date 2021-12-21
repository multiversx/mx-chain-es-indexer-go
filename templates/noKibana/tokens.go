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
				"type":   "date",
				"format": "epoch_second",
			},
			"nonce": Object{
				"type": "unsigned_long",
			},
			"data": Object{
				"type": "nested",
				"properties": Object{
					"name": Object{
						"type": "text",
					},
					"creator": Object{
						"type": "text",
					},
					"tags": Object{
						"type": "text",
					},
					"attributes": Object{
						"type": "text",
					},
					"metadata": Object{
						"type": "text",
					},
					"ownersHistory": Object{
						"type": "nested",
						"properties": Object{
							"timestamp": Object{
								"type":   "date",
								"format": "epoch_second",
							},
						},
					},
				},
			},
		},
	},
}
