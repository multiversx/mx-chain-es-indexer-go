package tags

import (
	"bytes"
)

// CountTags defines what a TagCount handler should be able to do
type CountTags interface {
	Serialize() ([]*bytes.Buffer, error)
	ParseTags(attributes []string)
	GetTags() []string
	Len() int
}

type tagsCount struct {
	tags map[string]int
}

// NewTagsCount will create a new instance of tagsCount, this structure is not concurrent save
func NewTagsCount() CountTags {
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
