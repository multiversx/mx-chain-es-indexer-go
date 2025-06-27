package indices

// TimestampMs holds the configuration for the timestampMs field
var TimestampMs = Object{
	"properties": Object{
		"timestampMs": Object{
			"type":   "date",
			"format": "epoch_millis",
		},
	},
}

// TokensTimestampMs holds the configuration for the tokens index timestampMs fields
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

// DelegatorsTimestampMs holds the configuration for the delegators index timestampMs fields
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

// DeploysTimestampMs holds the configuration for the scdeploys index timestampMs fields
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
			"properties": Object{
				"timestampMs": Object{
					"type":   "date",
					"format": "epoch_millis",
				},
			},
		},
	},
}
