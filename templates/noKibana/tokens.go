package noKibana

// Tokens will hold the configuration for the tokens index
var Tokens = Object{
	"index_patterns": Array{
		"tokens-*",
	},
	"template": Object{
		"settings": Object{
			"number_of_shards":   3,
			"number_of_replicas": 0,
		},
		"mappings": Object{
			"properties": Object{
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
							"type": "long",
						},
						"tags": Object{
							"type": "keyword",
						},
						"uris": Object{
							"type": "keyword",
						},
						"whiteListedStorage": Object{
							"type": "boolean",
						},
					},
				},
				"identifier": Object{
					"type": "text",
				},
				"issuer": Object{
					"type": "keyword",
				},
				"name": Object{
					"type": "keyword",
				},

				"nonce": Object{
					"type": "double",
				},
				"numDecimals": Object{
					"type": "long",
				},
				"ownersHistory": Object{
					"type": "nested",
					"properties": Object{
						"address": Object{
							"type": "keyword",
						},
						"timestamp": Object{
							"index":  "false",
							"type":   "date",
							"format": "epoch_second",
						},
					},
				},
				"properties": Object{
					"properties": Object{
						"canAddSpecialRoles": Object{
							"index": "false",
							"type":  "boolean",
						},
						"canBurn": Object{
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
						"canFreeze": Object{
							"index": "false",
							"type":  "boolean",
						},
						"canMint": Object{
							"index": "false",
							"type":  "boolean",
						},
						"canPause": Object{
							"index": "false",
							"type":  "boolean",
						},
						"canTransferNFTCreateRole": Object{
							"index": "false",
							"type":  "boolean",
						},
						"canUpgrade": Object{
							"index": "false",
							"type":  "boolean",
						},
						"canWipe": Object{
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
				"ticker": Object{
					"type": "keyword",
				},
				"timestamp": Object{
					"type":   "date",
					"format": "epoch_second",
				},
				"changedToDynamicTimestamp": Object{
					"type":   "date",
					"format": "epoch_second",
				},
				"token": Object{
					"type": "keyword",
				},
				"type": Object{
					"type": "keyword",
				},
			},
		},
	},
}
