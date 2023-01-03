package withKibana

// Transactions will hold the configuration for the transactions index
var Transactions = Object{
	"index_patterns": Array{
		"transactions-*",
	},
	"settings": Object{
		"number_of_shards":   5,
		"number_of_replicas": 0,
		"opendistro.index_state_management.rollover_alias": "transactions",
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
			"data": Object{
				"type": "text",
			},
			"esdtValues": Object{
				"type": "keyword",
			},
			"fee": Object{
				"index": "false",
				"type":  "keyword",
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
			"gasUsed": Object{
				"index": "false",
				"type":  "long",
			},
			"hasOperations": Object{
				"type": "boolean",
			},
			"hasScResults": Object{
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
				"type": "long",
			},
			"operation": Object{
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
			"round": Object{
				"type": "long",
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
				"type": "keyword",
			},
			"value": Object{
				"type": "keyword",
			},
			"version": Object{
				"type": "long",
			},
		},
	},
}
