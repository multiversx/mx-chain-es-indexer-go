package tags

import "github.com/ElrondNetwork/elastic-indexer-go/data"

type tagsCount struct {
	tags map[string]int
}

// NewTagsCount will create a new instance of tagsCount, this structure is not concurrent save
func NewTagsCount() data.CountTags {
	return &tagsCount{
		tags: make(map[string]int),
	}
}

// Len will return the number of tags
func (tc *tagsCount) Len() int {
	return len(tc.tags)
}

// GetTags will return all the tags
func (tc *tagsCount) GetTags() []string {
	tags := make([]string, 0, tc.Len())
	for key := range tc.tags {
		tags = append(tags, key)
	}

	return tags
}

// ParseTags will parse all the tags
func (tc *tagsCount) ParseTags(tags []string) {
	if tags == nil {
		return
	}

	newTags := removeDuplicatedTags(tags)
	for _, tag := range newTags {
		if tag == "" {
			continue
		}

		tc.tags[tag]++
	}
}

func removeDuplicatedTags(stringsSlice []string) []string {
	keys := make(map[string]bool)
	list := make([]string, 0)

	for _, entry := range stringsSlice {
		_, exists := keys[entry]
		if exists {
			continue
		}

		keys[entry] = true
		list = append(list, entry)
	}
	return list
}
