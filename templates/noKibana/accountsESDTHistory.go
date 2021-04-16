package noKibana

// AccountsESDTHistory will hold the configuration for the accountsesdthistory index
var AccountsESDTHistory = Object{
	"index_patterns": Array{
		"accountsesdthistory-*",
	},
	"settings": Object{
		"number_of_shards":   5,
		"number_of_replicas": 0,
	},
}
