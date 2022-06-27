package converters

import "encoding/json"

// JsonEscape will format the provided string in a json compatible string
func JsonEscape(i string) string {
	b, err := json.Marshal(i)
	if err != nil {
		return ""
	}

	// Trim the beginning and trailing " character
	return string(b[1 : len(b)-1])
}
