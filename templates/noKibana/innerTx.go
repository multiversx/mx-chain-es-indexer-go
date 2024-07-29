package noKibana

// InnerTxs will hold the configuration for the inner transaction structure
var InnerTxs = Object{
	"properties": Object{
		"innerTransactions": Object{
			"properties": Object{
				"nonce": Object{
					"type": "double",
				},
				"value": Object{
					"type": "keyword",
				},
				"sender": Object{
					"type": "keyword",
				},
				"receiver": Object{
					"type": "keyword",
				},
				"senderUserName": Object{
					"type": "keyword",
				},
				"receiverUsername": Object{
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
				"data": Object{
					"type": "text",
				},
				"signature": Object{
					"index": "false",
					"type":  "keyword",
				},
				"chainID": Object{
					"type": "keyword",
				},
				"version": Object{
					"type": "long",
				},
				"options": Object{
					"type": "long",
				},
				"guardian": Object{
					"type": "keyword",
				},
				"guardianSignature": Object{
					"index": "false",
					"type":  "keyword",
				},
				"relayer": Object{
					"type": "keyword",
				},
			},
		},
	},
}
