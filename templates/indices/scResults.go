package indices

// SCResults will hold the configuration for the scresults index
var SCResults = Object{
	"index_patterns": Array{
		"scresults-*",
	},
	"template": Object{
		"settings": Object{
			"number_of_shards":   3,
			"number_of_replicas": 0,
			"index": Object{
				"sort.field": Array{
					"timestamp",
				},
				"sort.order": Array{
					"desc",
				},
			},
		},
		"mappings": Object{
			"properties": Object{
				"callType": Object{
					"type": "keyword",
				},
				"code": Object{
					"index": "false",
					"type":  "text",
				},
				"data": Object{
					"type": "text",
				},
				"esdtValues": Object{
					"type": "keyword",
				},
				"esdtValuesNum": Object{
					"type": "double",
				},
				"function": Object{
					"type": "keyword",
				},
				"gasLimit": Object{
					"index": "false",
					"type":  "double",
				},
				"gasPrice": Object{
					"index": "false",
					"type":  "double",
				},
				"hasOperations": Object{
					"type": "boolean",
				},
				"miniBlockHash": Object{
					"type": "keyword",
				},
				"nonce": Object{
					"type": "double",
				},
				"operation": Object{
					"type": "keyword",
				},
				"originalSender": Object{
					"type": "keyword",
				},
				"originalTxHash": Object{
					"type": "keyword",
				},
				"prevTxHash": Object{
					"type": "keyword",
				},
				"receiver": Object{
					"type": "keyword",
				},
				"receiverShard": Object{
					"type": "long",
				},
				"receivers": Object{
					"type": "keyword",
				},
				"receiversShardIDs": Object{
					"type": "long",
				},
				"relayedValue": Object{
					"index": "false",
					"type":  "keyword",
				},
				"relayerAddr": Object{
					"type": "keyword",
				},
				"returnMessage": Object{
					"type": "text",
				},
				"sender": Object{
					"type": "keyword",
				},
				"senderShard": Object{
					"type": "long",
				},
				"timestamp": Object{
					"type":   "date",
					"format": "epoch_second",
				},
				"tokens": Object{
					"type": "text",
				},
				"value": Object{
					"type": "keyword",
				},
				"valueNum": Object{
					"type": "double",
				},
			},
		},
	},
}
