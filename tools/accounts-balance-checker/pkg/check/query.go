package check

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type object = map[string]interface{}

const matchAllQuery = `{ "query": { "match_all": { } } }`

func encodeQuery(query object) (bytes.Buffer, error) {
	var buff bytes.Buffer
	if err := json.NewEncoder(&buff).Encode(query); err != nil {
		return bytes.Buffer{}, fmt.Errorf("error encoding query: %s", err.Error())
	}

	return buff, nil
}

func getDocumentsByIDsQuery(hashes []string, withSource bool) object {
	interfaceSlice := make([]string, 0, len(hashes))
	for idx := range hashes {
		interfaceSlice = append(interfaceSlice, hashes[idx])
	}

	return object{
		"query": object{
			"ids": object{
				"values": interfaceSlice,
			},
		},
		"_source": withSource,
	}
}

func getBalancesByAddress(addr string) object {
	return object{
		"query": object{
			"match": object{
				"address": addr,
			},
		},
	}
}

func queryGetLastTxForToken(identifier, addr string) *bytes.Buffer {
	queryBytes := fmt.Sprintf(`{
	"query": {
		"bool": {
			"must": [
				{
					"match": {
						"tokens": {
							"query":"%s",
							"operator":"AND"
						}
					}
				},
				{
					"match": {
						"sender": {
							"query":"%s",
							"operator":"AND"
						}
					}
				}
			]
		}
	},
	"sort": [
		{
			"timestamp": {
				"order":"desc"
			}
		}
	]
}`, identifier, addr)

	return bytes.NewBuffer([]byte(queryBytes))
}

func queryGetLastOperationForAddress(addr string) *bytes.Buffer {
	queryBytes := fmt.Sprintf(`{
	"query": {
		"bool": {
			"should": [
				{
					"match": {
						"sender": {
							"query":"%s",
							"operator":"AND"
						}
					}
				},
				{
					"match": {
						"receiver": {
							"query":"%s",
							"operator":"AND"
						}
					}
				}
			]
		}
	},
	"sort": [
		{
			"timestamp": {
				"order":"desc"
			}
		}
	]
}`, addr, addr)

	return bytes.NewBuffer([]byte(queryBytes))
}
