package noKibana

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
				"type": "text",
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
				"properties": Object{
					"ESDTRoleLocalBurn": Object{
						"type": "keyword",
					},
					"ESDTRoleLocalMint": Object{
						"type": "keyword",
					},
					"ESDTRoleNFTAddQuantity": Object{
						"type": "keyword",
					},
					"ESDTRoleNFTAddURI": Object{
						"type": "keyword",
					},
					"ESDTRoleNFTBurn": Object{
						"type": "keyword",
					},
					"ESDTRoleNFTCreate": Object{
						"type": "keyword",
					},
					"ESDTRoleNFTUpdateAttributes": Object{
						"type": "keyword",
					},
					"ESDTTransferRole": Object{
						"type": "keyword",
					},
				},
			},
		},
	},
}
