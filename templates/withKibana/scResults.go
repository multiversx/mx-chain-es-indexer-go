package withKibana

// SCResults will hold the configuration for the scresults index
var SCResults = Object{
	"index_patterns": Array{
		"scresults-*",
	},
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
		"opendistro.index_state_management.rollover_alias": "scresults",
	},
	"mappings": Object{
		"properties": Object{
			"callType": Object{
				"type": "keyword",
			},
			"code": Object{
				"index": "false",
				"type":  "keyword",
			},
			"data": Object{
				"type": "keyword",
			},
			"esdtValues": Object{
				"type": "keyword",
			},
			"function": Object{
				"type": "keyword",
			},
			"gasLimit": Object{
				"index": "false",
				"type":  "long",
			},
			"gasPrice": Object{
				"index": "false",
				"type":  "long",
			},
			"hasOperations": Object{
				"type": "boolean",
			},
			"miniBlockHash": Object{
				"type": "keyword",
			},
			"nonce": Object{
				"type": "long",
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
				"type": "keyword",
			},
			"value": Object{
				"type": "keyword",
			},
		},
	},
}
