package tags

import (
	"encoding/base64"
)

func (tc *tagsCount) TagsCountToPostgres() (map[string]int, error) {
	tags := make(map[string]int)
	for tag, count := range tc.tags {
		if tag == "" {
			continue
		}

		base64Tag := base64.StdEncoding.EncodeToString([]byte(tag))
		tags[base64Tag] = count
	}

	return tags, nil
}
