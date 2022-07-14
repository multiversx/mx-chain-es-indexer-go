package converters

import (
	"encoding/json"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

const defaultStr = "default"

var log = logger.GetOrCreate("indexer/converters")

// JsonEscape will format the provided string in a json compatible string
func JsonEscape(i string) string {
	b, err := json.Marshal(i)
	if err != nil {
		log.Warn("converters.JsonEscape something went wrong",
			"input", i,
			"error", err,
		)
		return defaultStr
	}

	// Trim the beginning and trailing " character
	return string(b[1 : len(b)-1])
}
