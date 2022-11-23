package withKibana

// ESDTs will hold the configuration for the esdts index
var ESDTs = Object{
	"index_patterns": Array{
		"esdts-*",
	},
	"settings": Object{
		"number_of_shards":   3,
		"number_of_replicas": 0,
	},
	"mappings": Object{
		"properties": Object{
			"name": Object{
				"type": "keyword",
			},
			"ticker": Object{
				"type": "keyword",
			},
			"token": Object{
				"type": "keyword",
			},
			"issuer": Object{
				"type": "keyword",
			},
			"currentOwner": Object{
				"type": "keyword",
			},
			"numDecimals": Object{
				"type": "long",
			},
			"type": Object{
				"type": "keyword",
			},
			"timestamp": Object{
				"type":   "date",
				"format": "epoch_second",
			},
			"data": Object{
				"type": "nested",
				"properties": Object{
					"name": Object{
						"index": "false",
						"type":  "keyword",
					},
					"creator": Object{
						"type": "keyword",
					},
					"tags": Object{
						"index": "false",
						"type":  "keyword",
					},
					"attributes": Object{
						"index": "false",
						"type":  "keyword",
					},
					"metadata": Object{
						"index": "false",
						"type":  "keyword",
					},
					"ownersHistory": Object{
						"type": "nested",
						"properties": Object{
							"timestamp": Object{
								"index":  "false",
								"type":   "date",
								"format": "epoch_second",
							},
							"address": Object{
								"type": "keyword",
							},
						},
					},
				},
			},
			"properties": Object{
				"properties": Object{
					"canMint": Object{
						"index": "false",
						"type":  "boolean",
					},
					"canBurn": Object{
						"index": "false",
						"type":  "boolean",
					},
					"canUpgrade": Object{
						"index": "false",
						"type":  "boolean",
					},
					"canTransferNFTCreateRole": Object{
						"index": "false",
						"type":  "boolean",
					},
					"canAddSpecialRoles": Object{
						"index": "false",
						"type":  "boolean",
					},
					"canPause": Object{
						"index": "false",
						"type":  "boolean",
					},
					"canFreeze": Object{
						"index": "false",
						"type":  "boolean",
					},
					"canWipe": Object{
						"index": "false",
						"type":  "boolean",
					},
					"canChangeOwner": Object{
						"index": "false",
						"type":  "boolean",
					},
					"canCreateMultiShard": Object{
						"index": "false",
						"type":  "boolean",
					},
				},
			},
			"roles": Object{
				"type": "nested",
			},
		},
	},
}
