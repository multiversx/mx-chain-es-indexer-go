package withKibana

// AccountsESDT will hold the configuration for the accountsesdt index
var AccountsESDT = Object{
	"index_patterns": Array{
		"accountsesdt-*",
	},
	"settings": Object{
		"number_of_shards":   3,
		"number_of_replicas": 0,
	},
	"mappings": Object{
		"properties": Object{
			"balanceNum": Object{
				"type": "double",
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
				},
			},
			"tokenNonce": Object{
				"type": "double",
			},
		},
	},
}
