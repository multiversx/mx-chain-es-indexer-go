package client

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func encode(obj objectsMap) (bytes.Buffer, error) {
	var buff bytes.Buffer
	if err := json.NewEncoder(&buff).Encode(obj); err != nil {
		return bytes.Buffer{}, fmt.Errorf("error encoding : %w", err)
	}

	return buff, nil
}

func getDocumentsByIDsQuery(hashes []string, withSource bool) objectsMap {
	interfaceSlice := make([]interface{}, len(hashes))
	for idx := range hashes {
		interfaceSlice[idx] = objectsMap{
			"_id":     hashes[idx],
			"_source": withSource,
		}
	}

	return objectsMap{
		"docs": interfaceSlice,
	}
}
