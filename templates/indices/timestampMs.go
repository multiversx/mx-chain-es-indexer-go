package indices

// TimestampMs will hold the configuration for the timestampMs field
var TimestampMs = Object{
	"properties": Object{
		"timestampMs": Object{
			"type":   "date",
			"format": "epoch_millis",
		},
	},
}

var TokensTimestampMs = Object{
	"properties": Object{
		"timestampMs": Object{
			"type":   "date",
			"format": "epoch_millis",
		},
		"changedToDynamicTimestampMs": Object{
			"type":   "date",
			"format": "epoch_millis",
		},
		"ownersHistory": Object{
			"type": "nested",
			"properties": Object{
				"timestampMs": Object{
					"index":  "false",
					"type":   "date",
					"format": "epoch_millis",
				},
			},
		},
	},
}

var DelegatorsTimestampMs = Object{
	"properties": Object{
		"timestampMs": Object{
			"type":   "date",
			"format": "epoch_millis",
		},
		"unDelegateInfo": Object{
			"properties": Object{
				"timestampMs": Object{
					"index":  "false",
					"type":   "date",
					"format": "epoch_millis",
				},
			},
		},
	},
}

var DeploysTimestampMs = Object{
	"properties": Object{
		"timestampMs": Object{
			"type":   "date",
			"format": "epoch_millis",
		},
		"upgrades": Object{
			"type": "nested",
			"properties": Object{
				"timestampMs": Object{
					"type":   "date",
					"format": "epoch_millis",
				},
			},
		},
		"owners": Object{
			"type": "nested",
			"properties": Object{
				"timestampMs": Object{
					"type":   "date",
					"format": "epoch_millis",
				},
			},
		},
	},
}
