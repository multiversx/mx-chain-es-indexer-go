package withKibana

// MiniblocksPolicy will hold the configuration for the miniblocks index policy
var MiniblocksPolicy = Object{
	"policy": Object{
		"description":   "Open distro policy for the miniblocks elastic index.",
		"default_state": "hot",
		"states": Array{
			Object{
				"name": "hot",
				"actions": Array{
					Object{
						"rollover": Object{
							"min_size": "60gb",
						},
					},
				},
				"transitions": Array{
					Object{
						"state_name": "warm",
						"conditions": Object{
							"min_size": "60gb",
						},
					},
				},
			},
			Object{
				"name": "warm",
				"actions": Array{
					Object{
						"replica_count": Object{
							"number_of_replicas": 1,
						},
					},
				},
				"transitions": Array{},
			},
		},
		"ism_template": Object{
			"index_patterns": Array{"miniblocks-*"},
			"priority":       100,
		},
	},
}
