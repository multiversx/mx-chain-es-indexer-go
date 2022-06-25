package tags

import (
	"encoding/base64"
	"fmt"
	"regexp"

	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

// Serialize will serialize tagsCount in a way that Elastic Search expects a bulk request
func (tc *tagsCount) Serialize(buffSlice *data.BufferSlice, index string) error {
	for tag, count := range tc.tags {
		if tag == "" {
			continue
		}

		isAlphanumeric := regexp.MustCompile(`^[a-zA-Z\d]*$`).MatchString(tag)
		if !isAlphanumeric {
			continue
		}

		base64Tag := base64.StdEncoding.EncodeToString([]byte(tag))
		meta := []byte(fmt.Sprintf(`{ "update" : {"_index":"%s", "_id" : "%s" } }%s`, index, base64Tag, "\n"))

		codeToExecute := `
			ctx._source.count += params.count; 
			ctx._source.tag = params.tag
`
		serializedDataStr := fmt.Sprintf(`{"script": {"source": "%s","lang": "painless","params": {"count": %d, "tag": "%s"}},"upsert": {"count": %d, "tag":"%s"}}`,
			converters.FormatPainlessSource(codeToExecute), count, tag, count, tag,
		)

		err := buffSlice.PutData(meta, []byte(serializedDataStr))
		if err != nil {
			return err
		}
	}

	return nil
}
