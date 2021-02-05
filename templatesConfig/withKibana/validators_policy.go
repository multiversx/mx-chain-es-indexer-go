package withKibana

var ValidatorsPolicy = Object{
	"policy": Object{
		"description":   "Open distro policy for the validators elastic index.",
		"default_state": "hot",
		"states": Array{
			Object{
				"name": "hot",
				"actions": Array{
					Object{
						"rollover": Object{
							"min_size": "20gb",
						},
					},
				},
				"transitions": Array{
					Object{
						"state_name": "warm",
						"conditions": Object{
							"min_size": "20gb",
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
	},
}
