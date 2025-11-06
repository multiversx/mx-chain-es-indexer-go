package indices

// ExecutionResults will hold the configuration for the execution results index
var ExecutionResults = Object{
	"index_patterns": Array{
		"executionresults-*",
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
				"round": Object{
					"type": "double",
				},
				"rootHash": Object{
					"index": "false",
					"type":  "keyword",
				},
				"notarizedInBlockHash": Object{
					"type": "keyword",
				},
				"epoch": Object{
					"type": "long",
				},
				"gasUsed": Object{
					"type": "double",
				},
				"txCount": Object{
					"index": "false",
					"type":  "long",
				},
				"accumulatedFees": Object{
					"index": "false",
					"type":  "keyword",
				},
				"developerFees": Object{
					"index": "false",
					"type":  "keyword",
				},
			},
		},
	},
}
