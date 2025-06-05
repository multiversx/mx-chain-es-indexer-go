package noKibana

// Operations will hold the configuration for the operations index
var Operations = Object{
	"index_patterns": Array{
		"operations-*",
	},
	"template": Object{
		"settings": Object{
			"number_of_shards":   5,
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
				"callType": Object{
					"index": "false",
					"type":  "keyword",
				},
				"canBeIgnored": Object{
					"type": "boolean",
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
				"fee": Object{
					"index": "false",
					"type":  "keyword",
				},
				"feeNum": Object{
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
				"gasUsed": Object{
					"index": "false",
					"type":  "double",
				},
				"hasOperations": Object{
					"type": "boolean",
				},
				"hasScResults": Object{
					"type": "boolean",
				},
				"hasLogs": Object{
					"type": "boolean",
				},
				"initialPaidFee": Object{
					"index": "false",
					"type":  "keyword",
				},
				"isRelayed": Object{
					"type": "boolean",
				},
				"isScCall": Object{
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
				"round": Object{
					"type": "double",
				},
				"searchOrder": Object{
					"type": "long",
				},
				"sender": Object{
					"type": "keyword",
				},
				"senderShard": Object{
					"type": "long",
				},
				"senderUserName": Object{
					"type": "keyword",
				},
				"signature": Object{
					"index": "false",
					"type":  "keyword",
				},
				"status": Object{
					"type": "keyword",
				},
				"timestamp": Object{
					"type":   "date",
					"format": "epoch_second",
				},
				"tokens": Object{
					"type": "text",
				},
				"type": Object{
					"type": "keyword",
				},
				"value": Object{
					"type": "keyword",
				},
				"valueNum": Object{
					"type": "double",
				},
				"version": Object{
					"type": "long",
				},
				"guardian": Object{
					"type": "keyword",
				},
				"guardianSignature": Object{
					"index": "false",
					"type":  "keyword",
				},
			},
		},
	},
}
