package split

import "fmt"

const searchSIZE = 9999

func computeQueryBasedOnTimestamp(timestamp uint64) []byte {
	if timestamp == 0 {
		query := fmt.Sprintf(`{
  "size": %d,
  "query": {
    "match_all": {}
  },
  "sort": [
    {"timestamp": "asc"}
  ]
}`, searchSIZE)

		return []byte(query)
	}

	query := fmt.Sprintf(`
{
  "size": %d,
  "query": {
    "match_all": {}
  },
  "sort": [
    {"timestamp": "asc"}
  ],
  "search_after": ["%d"]
}
`, searchSIZE, timestamp)

	return []byte(query)
}
