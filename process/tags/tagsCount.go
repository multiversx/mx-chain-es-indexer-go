package tags

import (
	"bytes"
	"encoding/json"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

const (
	tagsKey = "tags"
)

// CountTags defines what a TagCount handler should be able to do
type CountTags interface {
	Serialize() ([]*bytes.Buffer, error)
	ParseTagsFromDB(response map[string]interface{}) error
	ExtractTagsFromAttributes(attributes *data.Attributes)
	GetTags() []string
	Len() int
}

type tagsCount struct {
	tags map[string]int
}

// NewTagsCount will create a new instance of tagsCount
func NewTagsCount() CountTags {
	return &tagsCount{
		tags: make(map[string]int),
	}
}

func (tc *tagsCount) Len() int {
	return len(tc.tags)
}

func (tc *tagsCount) GetTags() []string {
	tags := make([]string, 0, tc.Len())
	for key := range tc.tags {
		tags = append(tags, key)
	}

	return tags
}

func (tc *tagsCount) ExtractTagsFromAttributes(attributes *data.Attributes) {
	tagsC := make(map[string]int)

	if attributes == nil {
		return
	}

	for key, tags := range *attributes {
		if key != tagsKey {
			continue
		}

		for _, tag := range tags {
			tagsC[tag] = 1
		}

	}

	for key, value := range tagsC {
		_, ok := tc.tags[key]
		if ok {
			tc.tags[key] += value
			continue
		}

		tc.tags[key] = value
	}
}

func (tc *tagsCount) ParseTagsFromDB(response map[string]interface{}) error {
	if response == nil {
		return nil
	}

	responseDecoded, err := getResponse(response)
	if err != nil {
		return err
	}

	for _, tagRes := range responseDecoded.Docs {
		if !tagRes.Found {
			continue
		}

		count, ok := tc.tags[tagRes.Tag]
		if !ok {
			continue
		}

		tc.tags[tagRes.Tag] = count + tagRes.Source.Count
	}

	return nil
}

func getResponse(response map[string]interface{}) (*responseTags, error) {
	resBytes, err := json.Marshal(&response)
	if err != nil {
		return nil, err
	}

	res := &responseTags{}
	err = json.Unmarshal(resBytes, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}
