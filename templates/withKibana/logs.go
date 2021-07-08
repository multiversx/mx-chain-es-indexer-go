package withKibana

// Logs will hold the configuration for the logs index
var Logs = Object{
	"index_patterns": Array{
		"logs-*",
	},
	"settings": Object{
		"number_of_shards":   3,
		"number_of_replicas": 0,
	},
	"mappings": Object{
		"properties": Object{
			"address": Object{
				"type": "text",
			},
			"events": Object{
				"type": "nested",
				"properties": Object{
					"address": Object{
						"type": "text",
					},
					"topics": Object{
						"type": "text",
					},
					"data": Object{
						"type": "text",
					},
				},
			},
		},
	},
}
