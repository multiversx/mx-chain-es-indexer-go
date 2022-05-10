package check

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type object = map[string]interface{}

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
