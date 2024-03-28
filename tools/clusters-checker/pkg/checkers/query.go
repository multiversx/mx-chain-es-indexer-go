package checkers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

type object = map[string]interface{}

func getAll(withSource bool) []byte {
	query := fmt.Sprintf(`{"query": {"match_all": {}},"_source": %s}`, strconv.FormatBool(withSource))

	return []byte(query)
}

func queryMultipleObj(ids []string, withSource bool) []byte {
	query := object{
		"query": object{
			"terms": object{
				"_id": ids,
			},
		},
		"_source": withSource,
	}

	var buff bytes.Buffer
	_ = json.NewEncoder(&buff).Encode(query)

	return buff.Bytes()
}

func getAllSortTimestampASC(withSource bool, start, stop int) []byte {
	obj := object{
		"query": object{
			"range": object{
				"timestamp": object{
					"gte": fmt.Sprintf("%d", start),
					"lte": fmt.Sprintf("%d", stop),
				},
			},
		},
		"_source": withSource,
		"sort": []interface{}{
			object{
				"timestamp": object{
					"order": "desc",
				},
			},
		},
	}

	var buff bytes.Buffer
	_ = json.NewEncoder(&buff).Encode(obj)

	return buff.Bytes()
}
