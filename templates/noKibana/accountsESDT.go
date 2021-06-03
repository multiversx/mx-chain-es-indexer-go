package noKibana

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
			"tokenMetaData": Object{
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
