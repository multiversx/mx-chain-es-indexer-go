package indices

// EpochInfo will hold the configuration for the epochinfo index
var EpochInfo = Object{
	"index_patterns": Array{
		"epochinfo-*",
	},
	"template": Object{
		"settings": Object{
			"number_of_shards":   3,
			"number_of_replicas": 0,
		},
		"mappings": Object{
			"properties": Object{
				"accumulatedFees": Object{
					"type": "keyword",
				},
				"developerFees": Object{
					"type": "keyword",
				},
			},
		},
	},
}
