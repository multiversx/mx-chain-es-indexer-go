package noKibana

// AccountsESDT will hold the configuration for the accountsesdt index
var AccountsESDT = Object{
	"index_patterns": Array{
		"accountsesdt-*",
	},
	"template": Object{
		"settings": Object{
			"number_of_shards":   3,
			"number_of_replicas": 0,
		},
		"mappings": Object{
			"properties": Object{
				"address": Object{
					"type": "keyword",
				},
				"balance": Object{
					"type": "keyword",
				},
				"balanceNum": Object{
					"type": "double",
				},
				"currentOwner": Object{
					"type": "keyword",
				},
				"data": Object{
					"type": "nested",
					"properties": Object{
						"attributes": Object{
							"index": "false",
							"type":  "keyword",
						},
						"creator": Object{
							"type": "keyword",
						},
						"hash": Object{
							"index": "false",
							"type":  "keyword",
						},
						"metadata": Object{
							"index": "false",
							"type":  "keyword",
						},
						"name": Object{
							"type": "keyword",
						},
						"nonEmptyURIs": Object{
							"type": "boolean",
						},
						"royalties": Object{
							"index": "false",
							"type":  "long",
						},
						"tags": Object{
							"type": "text",
						},
						"uris": Object{
							"type": "text",
						},
					},
				},
				"identifier": Object{
					"type": "text",
				},
				"properties": Object{
					"type": "keyword",
				},
				"shardID": Object{
					"type": "long",
				},
				"timestamp": Object{
					"type":   "date",
					"format": "epoch_second",
				},
				"token": Object{
					"type": "keyword",
				},
				"tokenNonce": Object{
					"type": "double",
				},
				"type": Object{
					"type": "keyword",
				},
			},
		},
	},
}
