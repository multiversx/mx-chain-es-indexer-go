package indices

// Blocks will hold the configuration for the blocks index
var Blocks = Object{
	"index_patterns": Array{
		"blocks-*",
	},
	"template": Object{
		"settings": Object{
			"number_of_shards":   3,
			"number_of_replicas": 0,
			"index": Object{
				"sort.field": Array{
					"timestamp", "nonce",
				},
				"sort.order": Array{
					"desc", "desc",
				},
			},
		},
		"mappings": Object{
			"properties": Object{
				"accumulatedFees": Object{
					"index": "false",
					"type":  "keyword",
				},
				"developerFees": Object{
					"index": "false",
					"type":  "keyword",
				},
				"epoch": Object{
					"type": "long",
				},
				"epochStartBlock": Object{
					"type": "boolean",
				},
				"epochStartInfo": Object{
					"properties": Object{
						"nodePrice": Object{
							"index": "false",
							"type":  "keyword",
						},
						"prevEpochStartHash": Object{
							"index": "false",
							"type":  "keyword",
						},
						"prevEpochStartRound": Object{
							"index": "false",
							"type":  "double",
						},
						"rewardsForProtocolSustainability": Object{
							"index": "false",
							"type":  "keyword",
						},
						"rewardsPerBlock": Object{
							"index": "false",
							"type":  "keyword",
						},
						"totalNewlyMinted": Object{
							"index": "false",
							"type":  "keyword",
						},
						"totalSupply": Object{
							"index": "false",
							"type":  "keyword",
						},
						"totalToDistribute": Object{
							"index": "false",
							"type":  "keyword",
						},
					},
				},
				"epochStartShardsData": Object{
					"properties": Object{
						"epoch": Object{
							"index": "false",
							"type":  "long",
						},
						"firstPendingMetaBlock": Object{
							"index": "false",
							"type":  "keyword",
						},
						"headerHash": Object{
							"index": "false",
							"type":  "keyword",
						},
						"lastFinishedMetaBlock": Object{
							"index": "false",
							"type":  "keyword",
						},
						"nonce": Object{
							"index": "false",
							"type":  "double",
						},
						"pendingMiniBlockHeaders": Object{
							"properties": Object{
								"hash": Object{
									"index": "false",
									"type":  "keyword",
								},
								"receiverShard": Object{
									"index": "false",
									"type":  "long",
								},
								"senderShard": Object{
									"index": "false",
									"type":  "long",
								},
								"timestamp": Object{
									"index":  "false",
									"type":   "date",
									"format": "epoch_second",
								},
								"type": Object{
									"index": "false",
									"type":  "keyword",
								},
							},
						},
						"rootHash": Object{
							"index": "false",
							"type":  "keyword",
						},
						"round": Object{
							"index": "false",
							"type":  "double",
						},
						"scheduledRootHash": Object{
							"index": "false",
							"type":  "keyword",
						},
						"shardID": Object{
							"index": "false",
							"type":  "long",
						},
					},
				},
				"gasPenalized": Object{
					"type": "double",
				},
				"gasProvided": Object{
					"type": "double",
				},
				"gasRefunded": Object{
					"type": "double",
				},
				"maxGasLimit": Object{
					"type": "double",
				},
				"miniBlocksDetails": Object{
					"properties": Object{
						"firstProcessedTx": Object{
							"index": "false",
							"type":  "long",
						},
						"lastProcessedTx": Object{
							"index": "false",
							"type":  "long",
						},
						"mbIndex": Object{
							"index": "false",
							"type":  "long",
						},
					},
				},
				"miniBlocksHashes": Object{
					"type": "keyword",
				},
				"nonce": Object{
					"type": "double",
				},
				"notarizedBlocksHashes": Object{
					"type": "keyword",
				},
				"notarizedTxsCount": Object{
					"index": "false",
					"type":  "long",
				},
				"prevHash": Object{
					"type": "keyword",
				},
				"proposer": Object{
					"type": "long",
				},
				"pubKeyBitmap": Object{
					"index": "false",
					"type":  "keyword",
				},
				"round": Object{
					"type": "double",
				},
				"scheduledData": Object{
					"properties": Object{
						"accumulatedFees": Object{
							"index": "false",
							"type":  "keyword",
						},
						"developerFees": Object{
							"index": "false",
							"type":  "keyword",
						},
						"gasProvided": Object{
							"index": "false",
							"type":  "double",
						},
						"gasRefunded": Object{
							"index": "false",
							"type":  "double",
						},
						"penalized": Object{
							"index": "false",
							"type":  "double",
						},
						"rootHash": Object{
							"index": "false",
							"type":  "keyword",
						},
					},
				},
				"searchOrder": Object{
					"type": "long",
				},
				"shardId": Object{
					"type": "long",
				},
				"size": Object{
					"index": "false",
					"type":  "long",
				},
				"sizeTxs": Object{
					"index": "false",
					"type":  "long",
				},
				"stateRootHash": Object{
					"type": "keyword",
				},
				"timestamp": Object{
					"type":   "date",
					"format": "epoch_second",
				},
				"txCount": Object{
					"index": "false",
					"type":  "long",
				},
				"validators": Object{
					"index": "false",
					"type":  "long",
				},
			},
		},
	},
}
