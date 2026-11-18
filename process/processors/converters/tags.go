package converters

import (
	"strings"
)

const (
	// MaxIDSize is the maximum size of a document id
	MaxIDSize = 512

	attributesSeparator = ";"
	keyValuesSeparator  = ":"
	valuesSeparator     = ","
	tagsKey             = "tags"
	metadataKey         = "metadata"
)

// ExtractTagsFromAttributes will extract tags from the attributes
func ExtractTagsFromAttributes(attributes []byte) []string {
	return extractFromAttributes(attributes, tagsKey)
}

// ExtractMetaDataFromAttributes will extract metadata from attributes
func ExtractMetaDataFromAttributes(attributes []byte) string {
	res := extractFromAttributes(attributes, metadataKey)
	if len(res) < 1 {
		return ""
	}

	return TruncateFieldIfExceedsMaxLength(res[0])
}

func extractFromAttributes(attributes []byte, key string) []string {
	if len(attributes) == 0 {
		return nil
	}

	sAttributes := strings.Split(string(attributes), attributesSeparator)

	for _, keValuesPair := range sAttributes {
		sKeyValuesPair := strings.Split(keValuesPair, keyValuesSeparator)
		if len(sKeyValuesPair) < 2 {
			continue
		}
		if sKeyValuesPair[0] != key {
			continue
		}

		tagsSplit := strings.Split(sKeyValuesPair[1], valuesSeparator)

		return extractNonEmpty(tagsSplit, key)
	}

	return nil
}

func extractNonEmpty(tags []string, key string) []string {
	nonEmptyTags := make([]string, 0)
	for _, tag := range tags {
		if tag == "" {
			continue
		}

		modifiedTag := tag
		if key == tagsKey {
			modifiedTag = strings.ToLower(tag)
		}

		nonEmptyTags = append(nonEmptyTags, modifiedTag)
	}

	if len(nonEmptyTags) == 0 {
		return nil
	}

	return nonEmptyTags
}
