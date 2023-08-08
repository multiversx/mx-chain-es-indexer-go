package noKibana

// SCDeploys will hold the configuration for the scdeploys index
var SCDeploys = Object{
	"index_patterns": Array{
		"scdeploys-*",
	},
	"settings": Object{
		"number_of_shards":   3,
		"number_of_replicas": 0,
	},
	"mappings": Object{
		"properties": Object{
			"deployTxHash": Object{
				"type": "keyword",
			},
			"deployer": Object{
				"type": "keyword",
			},
			"timestamp": Object{
				"type":   "date",
				"format": "epoch_second",
			},
			"currentOwner": Object{
				"type": "keyword",
			},
			"initialCodeHash": Object{
				"type": "keyword",
			},
			"upgrades": Object{
				"type": "nested",
				"properties": Object{
					"timestamp": Object{
						"type":   "date",
						"format": "epoch_second",
					},
					"upgradeTxHash": Object{
						"type": "keyword",
					},
					"upgrader": Object{
						"type": "keyword",
					},
					"codeHash": Object{
						"type": "keyword",
					},
				},
			},
			"owners": Object{
				"type": "nested",
				"properties": Object{
					"timestamp": Object{
						"type":   "date",
						"format": "epoch_second",
					},
					"txHash": Object{
						"type": "keyword",
					},
					"address": Object{
						"type": "keyword",
					},
				},
			},
		},
	},
}
